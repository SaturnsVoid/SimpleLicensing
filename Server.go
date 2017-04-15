package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"Go-Simple-Licensing-System/SimpleLicensing"

	"github.com/gorilla/mux"
	"github.com/pelletier/go-toml"
)

var (
	PORT string
	SSL  bool
	KEY  string

	HOST     string
	DBPORT   string
	DATABASE string
	USERNAME string
	PASSWORD string

	db  *sql.DB
	err error

	ConfigRaw string = `[server]
port = "{PORT}"
ssl = {SSL}
key = "{KEY}"

[database]
host = "{HOST}"
db = "{DB}"
username = "{USERNAME}"
password = "{PASSWORD}"`
)

func loadConfig() {
	config, err := toml.LoadFile("config.toml")
	if err != nil {
		fmt.Println("[ERROR] Could not load config.toml!")
		os.Exit(0)
	} else {
		PORT = config.Get("server.port").(string)
		SSL = config.Get("server.ssl").(bool)
		KEY = config.Get("server.key").(string)

		HOST = config.Get("database.host").(string)
		DATABASE = config.Get("database.db").(string)
		USERNAME = config.Get("database.username").(string)
		PASSWORD = config.Get("database.password").(string)
	}
}

func convertInt(input string) (bool, int) {
	i, err := strconv.Atoi(input)
	if err != nil {
		return false, 0
	}
	return true, i
}

func convertBool(input string) (bool, bool) {
	i, err := strconv.ParseBool(input)
	if err != nil {
		return false, false
	}
	return true, i
}

func randomString(n int) string {
	var letterRunes = []rune("1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func setup() {
	fmt.Println("Server Setup")
	fmt.Println("")
	var tmpconfig string = ConfigRaw

SetupPort:
	fmt.Print("Port # to Run Server On: ")
	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	a, _ := convertInt(scan.Text())
	if !a {
		fmt.Println("Port can only be numbers")
		goto SetupPort
	}
	tmpconfig = strings.Replace(tmpconfig, "{PORT}", scan.Text(), -1)
SetupSSL:
	fmt.Print("Use SSL (true/false): ")
	scan = bufio.NewScanner(os.Stdin)
	scan.Scan()
	c, _ := convertBool(scan.Text())
	if !c {
		fmt.Println("Must be true or false")
		goto SetupSSL
	}
	tmpconfig = strings.Replace(tmpconfig, "{SSL}", scan.Text(), -1)

	fmt.Println("Generating Secure Key...")
	tmpconfig = strings.Replace(tmpconfig, "{KEY}", randomString(16), -1)
	fmt.Print("SQL Database Host (127.0.0.1:3306): ")
	scan = bufio.NewScanner(os.Stdin)
	scan.Scan()
	tmpconfig = strings.Replace(tmpconfig, "{HOST}", scan.Text(), -1)

	fmt.Print("SQL Database: ")
	scan = bufio.NewScanner(os.Stdin)
	scan.Scan()
	tmpconfig = strings.Replace(tmpconfig, "{DB}", scan.Text(), -1)

	fmt.Print("SQL Database Username: ")
	scan = bufio.NewScanner(os.Stdin)
	scan.Scan()
	tmpconfig = strings.Replace(tmpconfig, "{USERNAME}", scan.Text(), -1)

	fmt.Print("SQL Database Password: ")
	scan = bufio.NewScanner(os.Stdin)
	scan.Scan()
	tmpconfig = strings.Replace(tmpconfig, "{PASSWORD}", scan.Text(), -1)

	fmt.Println("")
	fmt.Println("Saving config.toml file...")
	d1 := []byte(tmpconfig)
	err := ioutil.WriteFile("config.toml", d1, 0644)
	if err != nil {
		fmt.Println("There was an error creating the config file, please do it manually.")
	}
	if Licensing.CheckFileExist("config.toml") {
		fmt.Println("Setup Finished.")
	}
}

func indexHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "Go Simple Licensing Server")
}

func checkHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	license := request.FormValue("license")
	var tmpexp string
	decrypted := Licensing.Decrypt(KEY, license)
	err := db.QueryRow("SELECT experation FROM licenses WHERE license='" + decrypted + "'").Scan(&tmpexp)
	if err == sql.ErrNoRows { //No License for Key found
		fmt.Fprintf(response, "Bad.")
	} else { //Check Experation date
		ip := strings.Split(request.RemoteAddr, ":")[0]
		_, err := db.Exec("UPDATE licenses SET ip='" + ip + "' WHERE license='" + decrypted + "'")
		if err != nil {
			fmt.Println(err)
		}
		t, err := time.Parse("2006-01-02", tmpexp)
		if err != nil {
			fmt.Println("ERROR: SQL Table Date no Correct Format")
		}
		t2, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
		if t.After(t2) {
			fmt.Fprintf(response, "Good")
		} else {
			fmt.Fprintf(response, "Expired")
		}
	}
}

func API() {
	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/check", checkHandler).Methods("POST")
	http.Handle("/", router)
	if SSL {
		err := http.ListenAndServeTLS(":"+PORT, "server.crt", "server.key", nil) //:443
		if err != nil {
			fmt.Println("SSL Server Error: " + err.Error())
			os.Exit(0)
		}
	} else {
		http.ListenAndServe(":"+PORT, nil)
	}
}

func count() int { //Count Bot Rows
	rows, err := db.Query("SELECT COUNT(*) AS count FROM licenses")
	if err != nil {
		return 0
	}
	var count int

	defer rows.Close()
	for rows.Next() {
		rows.Scan(&count)
	}
	return count
}

func main() {
	fmt.Println("Go Simple Licensing System")

	if SSL {
		if !Licensing.CheckFileExist("server.crt") || !Licensing.CheckFileExist("server.key") {
			fmt.Println("[!] WARNING MAKE SURE YOU HAVE YOUR SSL FILES IN THE SAME DIR [!]")
			os.Exit(0)
		}
	}

	if !Licensing.CheckFileExist("config.toml") {
		setup()
	}

	loadConfig()

	db, err = sql.Open("mysql", USERNAME+":"+PASSWORD+"@tcp("+HOST+")/"+DATABASE)
	if err != nil {
		fmt.Println("[!] ERROR: CHECK MYSQL SETTINGS! [!]")
		os.Exit(0)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("[!] ERROR: CHECK IF MYSQL SERVER IS ONLINE! [!]")
		os.Exit(0)
	}

	go API()

	for {
	Menu:
		fmt.Println("[Total Licenses]", count())
		fmt.Println("")
		fmt.Print("Console: ")
		scan := bufio.NewScanner(os.Stdin)
		scan.Scan()
		switch scan.Text() {
		case "add":
			var email string
			var experation string
			var license string

			fmt.Print("License Email: ")
			scan = bufio.NewScanner(os.Stdin)
			scan.Scan()
			email = scan.Text()
		exp:
			fmt.Print("License Experation (YYYY-MM-DD): ")
			scan = bufio.NewScanner(os.Stdin)
			scan.Scan()
			_, err = time.Parse("2006-01-02", scan.Text())
			if err != nil {
				fmt.Println("Experation must be in the YYYY-MM-DD Format.")
				goto exp
			}
			experation = scan.Text()
			fmt.Println("")

			license = randomString(4) + "-" + randomString(4) + "-" + randomString(4)

			var tmpemail string
			err := db.QueryRow("SELECT email FROM licenses WHERE license='" + license + "'").Scan(&tmpemail)
			if err == sql.ErrNoRows {
				_, err = db.Exec("INSERT INTO licenses(email, license, experation, ip) VALUES(?, ?, ?, ?)", email, license, experation, "none")
				if err != nil {
					fmt.Println("[!] ERROR: UNABLE TO INSERT INTO DATABASE [!]")
					fmt.Println("")
					goto Menu
				}
			} else {
				fmt.Println("License already in database?")
				fmt.Println("License:", Licensing.Encrypt(KEY, license))
				fmt.Println("Email:", tmpemail)
				fmt.Println("")
				goto Menu
			}

			fmt.Println("License Key Generated!")
			fmt.Println("")
			fmt.Println("License Email:", email)
			fmt.Println("License Experation:", experation)
			fmt.Println("Save this as license.dat")
			fmt.Println("")
			fmt.Println(Licensing.Encrypt(KEY, license))
			fmt.Println("")

		case "remove":
			fmt.Print("License Email: ")
			scan = bufio.NewScanner(os.Stdin)
			scan.Scan()
			var tmp string
			err = db.QueryRow("SELECT license FROM licenses WHERE email=?", scan.Text()).Scan(&tmp)
			if err == sql.ErrNoRows {
				fmt.Println("[!] ERROR: COULD NOT FIND LICENSE [!]")
				fmt.Println("")
				goto Menu
			} else {
				fmt.Println("License Found:", tmp)
				_ = db.QueryRow("DELETE FROM licenses WHERE email=?", scan.Text())
				fmt.Println("License removed from database.")
				fmt.Println("")
				goto Menu
			}
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Unknown Command")
		}
	}
}
