package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
);
const Version = "1.0.0"


type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		mtx   sync.Mutex
		mtxs map[string]*sync.Mutex
		directory string
		log     Logger
	}
)

type Options struct {
	Logger
}

func new(directory string, options *Options) (*Driver, error) {
	directory = filepath.Clean(directory)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
	}

	driver := Driver{
		directory:     directory,
		mtxs: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	if _, err := os.Stat(directory); err == nil {
		opts.Logger.Debug("Using '%s' (db already exists)\n", directory)
		return &driver, nil
	}

	opts.Logger.Debug("Creating the database at '%s'...\n", directory)
	return &driver, os.MkdirAll(directory, 0755)
}

func (d *Driver) Write(collection, rsrc string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	if rsrc == "" {
		return fmt.Errorf("missing resource - unable to save record (no name)")
	}

	mtx := d.getOrCreateMutex(collection)
	mtx.Lock()
	defer mtx.Unlock()

	directory := filepath.Join(d.directory, collection)
	finalPath := filepath.Join(directory, rsrc+".json")
	tempPath := finalPath + ".tmp"

	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := os.Create(tempPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, finalPath)
}

func (d *Driver) Read(collection, rsrc string, v interface{}) error {

	if collection == "" {
		return fmt.Errorf("missing collection - unable to read")
	}

	if rsrc == "" {
		return fmt.Errorf("missing resource - unable to read record (no name)")
	}

	recs := filepath.Join(d.directory, collection, rsrc)

	if _, err := stat(recs); err != nil {
		return err
	}

	b, err := os.ReadFile(recs + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {

	if collection == "" {
		return nil, fmt.Errorf("missing collection - unable to read")
	}
	directory := filepath.Join(d.directory, collection)

	if _, err := stat(directory); err != nil {
		return nil, err
	}

	files, _ := os.ReadDir(directory)

	var recs []string

	for _, file := range files {
		b, err := os.ReadFile(filepath.Join(directory, file.Name()))
		if err != nil {
			return nil, err
		}

		recs = append(recs, string(b))
	}
	return recs, nil
}

func (d *Driver) Delete(collection, rsrc string) error {

	path := filepath.Join(collection, rsrc)
	mtx := d.getOrCreateMutex(collection)
	mtx.Lock()
	defer mtx.Unlock()

	directory := filepath.Join(d.directory, path)

	switch fileFind, err := stat(directory); {
	case fileFind == nil, err != nil:
		return fmt.Errorf("unable to find the directory named %v", path)

	case fileFind.Mode().IsDir():
		return os.RemoveAll(directory)

	case fileFind.Mode().IsRegular():
		return os.RemoveAll(directory + ".json")
	}
	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {

	d.mtx.Lock()
	defer d.mtx.Unlock()
	m, ok := d.mtxs[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mtxs[collection] = m
	}

	return m
}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}

// Golang cannot understand JSON on its own (Struct is what Golang understands) Golang is not a JS framwork right, which is why we have to do json.Number, it is Golang's way to identify.

type addy struct{
	City string;
	County string;
	State string;
	Zipcode json.Number; 
	Country string;
}

type Customer struct{
	Name string;
	Contact string;
	Age json.Number;
	Company string;
	Address addy;
}

//You have complete control here, so trust this is gonna save you from so many errors and issues

func main(){
	directory := "./";

	database, err := new(directory,nil);

	if err != nil {
		fmt.Println("Something went wrong", err);
	}

// Now here I am hard coding the values that we want to store in our database
// We can take these values from the front end
// We can take these values from the Cmd Line
// We can take these values from APIs also

// I am hard coding these values that we can store in the DB, just for the sake of using it in our database and understanding the working of it

//kind of an array of struct Customer
workers := []Customer{
	{"Sahil", "703999999", "20", "Geekmeo LLC", addy{"Ashburn", "Loudon", "VA","20148","USA"} },
	{"Suchita", "703888888", "20", "Geekmeo LLC", addy{"Persimmon", "Fairfax", "VA","20170","USA"} },
	{"Leo", "703555555", "23", "CodeCraft", addy{"McLean", "Fairfax", "VA","22101","USA"} },
	{"Nina", "703666666", "21", "ByteBusters", addy{"Vienna", "Fairfax", "VA","22180","USA"} },
	{"Aisha", "703888888", "22", "Technova Inc", addy{"Herndon", "Fairfax", "VA","20170","USA"} },
	{"Emma", "703444444", "25", "DataDynamos", addy{"Chantilly", "Fairfax", "VA","20151","USA"} },
}

//Now let's iterate through the employees
//strict

for _, val := range workers{
	database.Write("customers", val.Name, Customer{
		Name: val.Name,
		Contact: val.Contact ,
		Age: val.Age,
		Company: val.Company,
		Address: val.Address,

	})
}

allCustomers := []Customer{}

recs, err := database.ReadAll("customers");
if err != nil { 
fmt.Println("Oops! Something went wrong!", err)
}
fmt.Println(recs);

for _, f := range recs {
	workerFound := Customer{}
	if err := json.Unmarshal([]byte(f), &workerFound); err != nil {
		fmt.Println("Error", err)
	}
	allCustomers = append(allCustomers, workerFound)
}
fmt.Println((allCustomers))

// if err := database.Delete("customers", "Nina"); err != nil {      
	// 	fmt.Println("Error", err)
	// }
// This can be called through APIs also, but for now we can just use this way of calling it here.

if err := database.Delete("customers", ""); err != nil {
	fmt.Println("Error", err)
}


}