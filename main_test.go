package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/vladyslav2/bitcoin2sql/pkg/json2sql"
)

var a App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	recreateTables()

	code := m.Run()

	os.Exit(code)
}

func recreateTables() {

	fileBuff, err := ioutil.ReadFile("sql/01_block.sql")
	if err != nil {
		log.Fatalf("Cannot read sql/01_block.sql, %v", err)
	}
	sql := string(fileBuff)

	if _, err := a.DB.Exec(sql); err != nil {
		log.Fatalf("Cannot exec sql/01_block.sql as db query, %v", err)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestEmptyMainPage(t *testing.T) {

	// Empty tables
	if err := truncateTables(); err != nil {
		t.Error(err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array '[]'. Got '%s'", body)
	}
}

func truncateTables() error {
	_, err := a.DB.Exec("truncate block cascade")
	return err
}

func loadTestBlocks() ([]map[string]interface{}, error) {

	blocks, err := json2sql.ParseFile(a.DB, "fixtures/01_block.json")
	if err != nil {
		log.Fatalf("Cannot load fixture: %v", err)
		return nil, err
	}

	return blocks, nil
}

func loadTestTransaction() ([]map[string]interface{}, error) {

	// Read data from fixtures
	trans, err := json2sql.ParseFile(a.DB, "fixtures/01_transaction.json")
	if err != nil {
		log.Fatalf("Cannot load fixture: %v", err)
		return nil, err
	}

	return trans, nil
}

func TestMainPage(t *testing.T) {

	// Empty tables
	if err := truncateTables(); err != nil {
		t.Error(err)
	}

	// Load 5 test blocks
	testBlocks, err := json2sql.ParseFile(a.DB, "fixtures/01_block.json")
	if err != nil {
		t.Errorf("Cannot load fixture: %v", err)
		return
	}

	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	blocks := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(response.Body.Bytes(), &blocks); err != nil {
		t.Errorf("Cannot parse block response, %v", response.Body.String())
	}

	if len(blocks) != 5 {
		t.Errorf("Should be 5 blocks on main page, got: %d", len(blocks))
	}

	ref := reflect.ValueOf(testBlocks[0])
	keysReflect := ref.MapKeys()

	for _, k := range keysReflect {
		if blocks[2][k.String()] != testBlocks[2][k.String()] {
			t.Errorf("Wrong data on the main page expected %v, got: %v", testBlocks[2], blocks[2])
		}
	}
}

func TestWrongBlock(t *testing.T) {
	// We don't have this hash in database
	req, _ := http.NewRequest("GET", "/000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b123456789", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestBlock(t *testing.T) {

	// Empty tables
	if err := truncateTables(); err != nil {
		t.Error(err)
	}

	// Load 5 test blocks
	testBlocks, err := json2sql.ParseFile(a.DB, "fixtures/01_block.json")
	if err != nil {
		t.Errorf("Cannot load fixture: %v", err)
		return
	}

	// Load 5 test transactions
	testTransactions, err := json2sql.ParseFile(a.DB, "fixtures/02_transaction.json")
	if err != nil {
		t.Errorf("Cannot load fixture: %v", err)
		return
	}

	// Load 5 test transactions
	_, err = json2sql.ParseFile(a.DB, "fixtures/03_address.json")
	if err != nil {
		t.Errorf("Cannot load fixture: %v", err)
		return
	}

	// Load 5 test transactions
	_, err = json2sql.ParseFile(a.DB, "fixtures/04_txin.json")
	if err != nil {
		t.Errorf("Cannot load fixture: %v", err)
		return
	}

	// Load 5 test transactions
	_, err = json2sql.ParseFile(a.DB, "fixtures/05_txout.json")
	if err != nil {
		t.Errorf("Cannot load fixture: %v", err)
		return
	}

	// We have this hash in database
	req, _ := http.NewRequest("GET", "/000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var block map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &block); err != nil {
		t.Errorf("cannot unmarshal block: %s", response.Body.String())
	}

	ref := reflect.ValueOf(testBlocks[0])
	keysReflect := ref.MapKeys()

	for _, k := range keysReflect {
		s := strings.ToLower(k.String())
		if testBlocks[2][s] != block[s] {
			t.Errorf("Wrong data on the block, key: %s, expected %v, got: %v", s, testBlocks[2][s], block[s])
		}
	}

	fmt.Print(block["Transactions"])
	if testTransactions[1]["ID"] != block["Transactions"] {
		t.Errorf("Wrong data on the transaction, expected %v, got: %v", testTransactions[1], block["Transactions"])
	}
}
