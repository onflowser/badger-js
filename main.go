package badger

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"syscall/js"
)

// Await waits for the Promise to be resolved and returns the value
func Await(p js.Value) (js.Value, error) {
	resCh := make(chan js.Value)
	var then js.Func
	then = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resCh <- args[0]
		return nil
	})
	defer then.Release()

	errCh := make(chan error)
	var catch js.Func
	catch = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		errCh <- js.Error{args[0]}
		return nil
	})
	defer catch.Release()

	p.Call("then", then).Call("catch", catch)

	select {
	case res := <-resCh:
		return res, nil
	case err := <-errCh:
		return js.Undefined(), err
	}
}
func Set(name string, value string) {
	Await(js.Global().Call("storage_set", js.ValueOf(name), js.ValueOf(value)))
}

func Get(name string) string {
	v, _ := Await(js.Global().Call("storage_get", js.ValueOf(name)))
	return v.String()
}

func Clear() {
	Await(js.Global().Call("storage_clear"))
}

func Entries(prefix []byte) []*Item {
	v, _ := Await(js.Global().Call("storage_entries"))
	hexPrefix := hex.EncodeToString(prefix)
	items := make([]*Item, 0)
	fmt.Println(v.Length())
	for i := 0; i < v.Length(); i++ {

		if !strings.HasPrefix(v.Index(i).Index(0).String(), hexPrefix) {
			continue
		}

		key, _ := hex.DecodeString(v.Index(i).Index(0).String())
		value, _ := hex.DecodeString(v.Index(i).Index(1).String())
		items = append(items, &Item{key: key, value: value})

	}

	return items
}

type DB struct {
}

type Logger interface {
	Errorf(string, ...interface{})
	Warningf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

type Options struct {
	Logger          Logger
	Truncate        bool
	BypassLockGuard bool
	Dir             string
}

func (o Options) WithKeepL0InMemory(_ bool) Options {
	return o
}

type Iterator struct {
	prefix []byte
	index  int
	length int
	items  []*Item
}
type IteratorOptions struct {
	Prefix []byte
}

var DefaultIteratorOptions = IteratorOptions{}

type Txn struct {
}

func (txn *Txn) Set(key, val []byte) error {
	Set(hex.EncodeToString(key), hex.EncodeToString(val))
	return nil
}
func (db *DB) NewTransaction(update bool) *Txn {
	return &Txn{}
}

func DefaultOptions(path string) Options {
	return Options{}
}

func Open(opt Options) (db *DB, err error) {
	return &DB{}, nil
}

func (txn *Txn) Discard() {
}
func (txn *Txn) NewIterator(opt IteratorOptions) *Iterator {
	return &Iterator{prefix: opt.Prefix, index: 0, length: 0}
}
func (txn *Txn) Commit() error {
	return nil
}
func (txn *Txn) Get(key []byte) (item *Item, rerr error) {

	var r = Get(hex.EncodeToString(key))
	decodedByteArray, _ := hex.DecodeString(r)

	if decodedByteArray == nil || len(decodedByteArray) == 0 {
		return nil, ErrKeyNotFound
	}
	return &Item{key: key, value: decodedByteArray}, nil
}

type WriteBatch struct {
}

func (db *DB) NewWriteBatch() *WriteBatch {
	return &WriteBatch{}
}

func (db *DB) View(fn func(txn *Txn) error) error {
	return fn(&Txn{})
}

func (db *DB) Update(fn func(txn *Txn) error) error {
	return fn(&Txn{})
}

func (db *DB) Close() error {
	return nil
}

func (db *DB) Sync() error {
	return nil
}

func (db *DB) RLock() {
}
func (db *DB) RUnlock() {
}

type Item struct {
	key   []byte
	value []byte
}

func (it *Iterator) Item() *Item {
	return it.items[it.index]
}
func (it *Iterator) Valid() bool {
	return it.index < it.length
}
func (it *Iterator) Next() {
	it.index++
}
func (it *Iterator) Seek(key []byte) {
	it.index = 0
}

func (it *Iterator) Close() {
}
func (it *Iterator) Rewind() {
	it.index = 0
	it.items = Entries(it.prefix)
	it.length = len(it.items)
}

var (
	ErrKeyNotFound = errors.New("Key not found")
	ErrNoRewrite   = errors.New("Value log GC attempt didn't result in any cleanup")
)

func (item *Item) Key() []byte {
	return item.key
}

func (item *Item) Value(fn func(val []byte) error) error {
	return fn(item.value)
}

func (item *Item) ValueCopy(dst []byte) ([]byte, error) {
	return item.value, nil
}

func (item *Item) ValueSize() int64 {
	return int64(len(item.value))
}

func (db *DB) RunValueLogGC(discardRatio float64) error {
	return nil
}
