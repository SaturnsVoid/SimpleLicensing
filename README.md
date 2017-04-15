# SimpleLicensing
A Go Based Licensing System for Digital Rights Management

Included is a simple Server i built to handle the backend things, a database.sql file (you need to import into MySQL), the main package (License.go), and a example client show you how it could be used.

This is something i put together fast, but it works. Feel free to use it as you please.

# How it Works

1. Server creates a table for license email, license, experation date, and ip
2. Server generates a license key (XXXX-XXXX-XXXX)
3. Server encrypts it using your Servers key (Generated on setup) and server gives you it
4. You give client to buyer/user with license.dat file that contains the encrypted license
5. Client connects to license server, sending its license.dat
6. Server trys to decrypt liscense.dat and comapire with any in database
7. If Server finds the key it check to see if the license is expiered and update last IP
8. If not expired client runs, else it will tell them its expiered and closes

NOTE: The client must have an internet connection to the license server!

# Licensing.go
Licensing can be called at anytime in your program, Just import and select to settings for it. Its a simple one line command.

Licensing.CheckLicense("{LICESNESSERVER}", {USESSL}, {SILENT})

* {LICESNESSERVER} = http://127.0.0.1:8080
* {USESSL} = true or false to use SSL (Your server will need this setting too).
* {SILENT} = Show messages or not.

# Server
* Ability to add new licenses to database
* Config Setup Tool
* Ability to remove liceses from database

# Config.toml Format

`
[server]
port = "{PORT}"
ssl = {SSL}
key = "{KEY}"

[database]
host = "{HOST}"
db = "{DB}"
username = "{USERNAME}"
password = "{PASSWORD}"

`

# Packages Used
* github.com/gorilla/mux
* github.com/pelletier/go-toml
* github.com/go-sql-driver/mysql

# Other
Go is a amazing and powerful programming language. If you already haven't, check it out; https://golang.org/

# Donations
<img src="https://blockchain.info/Resources/buttons/donate_64.png"/>
<p align="center">Please Donate To Bitcoin Address: <b>1AEbR1utjaYu3SGtBKZCLJMRR5RS7Bp7eE</b></p>
