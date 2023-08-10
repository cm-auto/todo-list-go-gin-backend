package filedb

// if memory mode is added this should probably be renamed
// to prototype_db

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"todo-list-backend/src/models"
)

type Database struct {
	dir             string
	listCollection  *Collection[models.List]
	entryCollection *Collection[models.Entry]
}

func NewDatabase(dir string) (*Database, error) {
	listCollection, err := NewCollection[models.List]("list", dir)
	if err != nil {
		return nil, err
	}
	entryCollection, err := NewCollection[models.Entry]("entry", dir)
	if err != nil {
		return nil, err
	}
	return &Database{
		dir:             dir,
		listCollection:  listCollection,
		entryCollection: entryCollection,
	}, nil
}

func (db *Database) GetListCollection() *Collection[models.List] {
	return db.listCollection
}

func (db *Database) GetEntryCollection() *Collection[models.Entry] {
	return db.entryCollection
}

// TODO add mutex to prevent race condition
type DataContainer[T any] struct {
	count uint64
	data  []T
}

type dataContainerJson[T any] struct {
	Count uint64 `json:"count"`
	Data  []T    `json:"data"`
}

func (self dataContainerJson[T]) toDataContainer() DataContainer[T] {
	return DataContainer[T]{self.Count, self.Data}
}

func (self *DataContainer[T]) toDataContainerJson() dataContainerJson[T] {
	return dataContainerJson[T]{self.count, self.data}
}

type Collection[T any] struct {
	name          string
	directory     string
	dataContainer DataContainer[T]
}

func writeData(data interface{}, filename string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func readData(filename string, data interface{}) error {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileContent, data)
	if err != nil {
		return err
	}

	return nil
}

func NewCollection[T any](name string, directory string) (*Collection[T], error) {

	basePath := path.Join(directory, name)
	filename := fmt.Sprintf("%s.json", basePath)

	dataContainerJson := dataContainerJson[T]{}

	err := readData(filename, &dataContainerJson)
	if err != nil {
		return nil, err
	}
	dataContainer := dataContainerJson.toDataContainer()

	return &Collection[T]{
		name:          name,
		directory:     directory,
		dataContainer: dataContainer,
	}, nil
}

func (c *Collection[T]) getFilename() string {
	basePath := path.Join(c.directory, c.name)
	return fmt.Sprintf("%s.json", basePath)
}

func (c *Collection[T]) save() error {
	filename := c.getFilename()
	c.dataContainer.count = uint64(len(c.dataContainer.data))
	dataContainerJson := c.dataContainer.toDataContainerJson()
	return writeData(dataContainerJson, filename)
}

func (c *Collection[T]) FindOne(predicate func(*T) bool) *T {
	for _, data := range c.dataContainer.data {
		if predicate(&data) {
			return &data
		}
	}
	return nil
}

func (c *Collection[T]) Find(predicate func(*T) bool) []T {
	results := make([]T, 0)
	for _, data := range c.dataContainer.data {
		if predicate(&data) {
			results = append(results, data)
		}
	}
	return results
}

func (c *Collection[T]) GetAll() []T {
	return c.dataContainer.data
}

func (c *Collection[T]) Append(data T) error {
	c.dataContainer.count++
	c.dataContainer.data = append(c.dataContainer.data, data)
	// TODO if there was an error undo count++ and append
	return c.save()
}

func (c *Collection[T]) DeleteOne(predicate func(*T) bool) (*T, error) {
	for i, data := range c.dataContainer.data {
		if predicate(&data) {
			c.dataContainer.data = append(c.dataContainer.data[:i], c.dataContainer.data[i+1:]...)
			c.dataContainer.count--
			return &data, c.save()
		}
	}
	return nil, nil
}

func (c *Collection[T]) DeleteMany(predicate func(*T) bool) (uint64, error) {
	deletedCounter := 0
	// for i, data := range c.dataContainer.data {
	// 	if predicate(&data) {
	// 		c.dataContainer.data = append(c.dataContainer.data[:i], c.dataContainer.data[i+1:]...)
	// 		c.dataContainer.count--
	// 		deletedCounter++
	// 	}
	// }

	// we can't do the previous loop, since the slice is getting shorter
	// and thus the index would not match anymore
	for i := len(c.dataContainer.data) - 1; i >= 0; i-- {
		data := c.dataContainer.data[i]
		if predicate(&data) {
			c.dataContainer.data = append(c.dataContainer.data[:i], c.dataContainer.data[i+1:]...)
			deletedCounter++
		}
	}
	return uint64(deletedCounter), c.save()
}

func (c *Collection[T]) PatchOne(predicate func(*T) bool, patchFunc func(*T)) (*T, error) {
	for i, data := range c.dataContainer.data {
		if predicate(&data) {
			// data is just a copy, so we have to get a reference
			// to the element in the slice
			itemReference := &c.dataContainer.data[i]
			patchFunc(itemReference)
			// the reason why we are returning a copy
			// is so that the item in the slice can't be modified
			// after the function has finished
			copy := *itemReference
			return &copy, c.save()
		}
	}
	return nil, nil
}

func (c *Collection[T]) PutOne(predicate func(*T) bool, data T) error {
	for i, data := range c.dataContainer.data {
		if predicate(&data) {
			c.dataContainer.data[i] = data
			return c.save()
		}
	}
	return nil
}

func (c *Collection[T]) Count() uint64 {
	return c.dataContainer.count
}
