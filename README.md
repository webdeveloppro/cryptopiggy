# Cryptopiggy app

Bitcoin application where you can search by address hash and see transaction connected with address, amounts and final ballance.
Its extends bitcoin ledger with functionality which hard/perfomance costly to do. Like analyzing transactions/blocks and etc
Like, lets say you want to see all transaction from a hash <xxx> for 2014 yea

In additional it tried to predict how much dollars you address have base on the time address recieve/send bitcoins.


## Tech implimentation
Each blockchain entity is presented as a dedicated package. Each package contains structure which extends blockchain ledger

### Block
Block structure provides quick ways to find transaction and bitcoin price based on date when block was created

### Address
Address structure provides you a quick very (<100ms) to find any address in a ledger.
In additonal it gives ability to find all transactions connected with address

### Transactions
Transaction structure provide a convenient way to to convert txin/txout data to jsonb format, look up for transactions base on search parameters and bitcoin price per transaction


## Installation
You need to have converted version of the bitcoin blockchain (see https://github.com/webdeveloppro/btcd2sql)

Before, prepare your golang envoirement
```
git clone https://github.com/webdeveloppro/cryptopiggy
cd cryptopiggy
vi .env <-- setup database credentials 
go build
go run webapp <--to run RESTFUL http endpoints
go run wsapp <-- to run websocket push server
```
