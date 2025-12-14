package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

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
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)

const Version = "1.0.1"

type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := Driver{
		dir:     dir,
		log:     opts.Logger,
		mutexes: make(map[string]*sync.Mutex),
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &driver, nil
	}

	opts.Logger.Info("Creating new database at '%s'...\n", dir)

	return &driver, os.Mkdir(dir, 0755)

}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if resource == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tempPath := fnlPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))
	if err := ioutil.WriteFile(tempPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, fnlPath)

}

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if resource == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name cannot be empty")
	}

	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := ioutil.ReadDir(dir)

	var records []string
	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		records = append(records, string(b))
	}

	return records, nil

}

func (d *Driver) Delete(collection, resource string) error {

	if collection == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	path := filepath.Join(collection, resource)
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil && err != nil:
		return fmt.Errorf("resource '%s' does not exist in collection '%s'", resource, collection)
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)
	case fi.Mode().IsRegular():
		return os.Remove(dir + ".json")
	}

	return nil

}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {

	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	dir := "./"

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error creating DB:", err)
	}

	employee := []User{
		{
			Name:    "John Doe",
			Age:     "30",
			Contact: "123-456-7890",
			Company: "TechCorp",
			Address: Address{
				City:    "San Francisco",
				State:   "CA",
				Country: "USA",
				Pincode: "94105",
			},
		},
		{
			Name:    "Jane Smith",
			Age:     "28",
			Contact: "987-654-3210",
			Company: "Innovatech",
			Address: Address{
				City:    "New York",
				State:   "NY",
				Country: "USA",
				Pincode: "10001",
			},
		},
		{
			Name:    "Alice Johnson",
			Age:     "35",
			Contact: "555-123-4567",
			Company: "WebSolutions",
			Address: Address{
				City:    "Los Angeles",
				State:   "CA",
				Country: "USA",
				Pincode: "90001",
			},
		},
		{
			Name:    "Bob Brown",
			Age:     "40",
			Contact: "444-555-6666",
			Company: "DataAnalytics",
			Address: Address{
				City:    "Chicago",
				State:   "IL",
				Country: "USA",
				Pincode: "60601",
			},
		},
	}

	for _, value := range employee {
		db.Write("users", value.Name, User{
			Name:    value.Name,
			Age:     value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,
		})
	}

	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("Error reading records:", err)
	}

	fmt.Println("All User Records:", records)

	allusers := []User{}
	for _, record := range records {
		employeeFound := User{}
		err := json.Unmarshal([]byte(record), &employeeFound)
		if err != nil {
			fmt.Println("Error unmarshaling record:", err)
		}
		allusers = append(allusers, employeeFound)
	}

	fmt.Println("All Users Structs:", allusers)

	if err := db.Delete("users", "Alice Johnson"); err != nil {
		fmt.Println("Error deleting record:", err)
	}

	if err := db.Delete("users", ""); err != nil {
		fmt.Println("Error deleting all records:", err)
	}

}
