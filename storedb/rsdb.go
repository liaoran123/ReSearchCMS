package storedb

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	RsDB *rsdb
	// 建立一个共享进程池
	Batch = sync.Pool{
		New: func() any {
			return new(leveldb.Batch)
		},
	}
)

// rsdb struct 使用单例模式，确保只能有一个实例
type rsdb struct {
	Opts *opt.Options
	Db   *leveldb.DB
}

// 保证所有连接都是使用RsDB
func OpenDb(dbpath string) *rsdb {
	RsDB = &rsdb{
		Opts: &opt.Options{
			// 示例：最大打开文件数、写入缓存等可在此配置
		},
	}
	RsDB.InitDB(dbpath)
	return RsDB //公共db，即保证所有连接都是使用该唯一db
}

// InitDB 初始化数据库连接
// dbpath 支持相对路径和绝对路径，会自动转换为跨平台兼容格式
func (s *rsdb) InitDB(dbpath string) error {
	if s.Db != nil {
		// 数据库已经初始化，直接返回
		return nil
	}
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("获取当前工作目录失败:", err)
	}
	// 构建数据库目录的绝对路径
	// 在当前工作目录下创建rsdbdb文件夹
	absolutePath := filepath.Join(currentDir, dbpath)
	// 使用 filepath.Clean() 确保路径跨平台兼容
	// 它会将路径转换为当前操作系统的路径分隔符，并处理 . 和 ..
	cleanPath := filepath.Clean(absolutePath)
	db, err := leveldb.OpenFile(cleanPath, s.Opts)
	if err != nil {
		// 如果损坏可以尝试 RecoverFile
		if db, err = leveldb.RecoverFile(cleanPath, nil); err != nil {
			log.Fatal(err)
			return err
		}
	}
	s.Db = db
	return nil
}

// 获取迭代器
// 由于golang不支持变参，故para只能传递两个参数
func (s *rsdb) GetIterator(para ...[]byte) iterator.Iterator {
	var slice *util.Range
	plen := len(para)
	switch plen {
	case 0: //不提供参数即是全库扫描
		slice = nil
	case 1: //提供一个参数，即是前缀扫描
		slice = util.BytesPrefix(para[0])
	default: //默认取前两个范围扫描
		slice = &util.Range{Start: para[0], Limit: para[1]}
	}
	return s.Db.NewIterator(slice, nil)
}

// Close 关闭数据库连接
func (s *rsdb) Close() error {
	if s.Db != nil {
		return s.Db.Close()
	}
	return nil
}
