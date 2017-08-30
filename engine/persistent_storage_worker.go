package engine

import (
	"bytes"
	"encoding/gob"
	"github.com/qiukeren/wukong/types"
	"sync/atomic"
)

type persistentStorageIndexDocumentRequest struct {
	docId string
	data  types.DocumentIndexData
}

func (engine *Engine) persistentStorageIndexDocumentWorker(shard int) {
	for {
		request := <-engine.persistentStorageIndexDocumentChannels[shard]

		// 得到value
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(request.data)
		if err != nil {
			atomic.AddUint64(&engine.numDocumentsStored, 1)
			continue
		}
		b := []byte(request.docId)
		// 将key-value写入数据库
		engine.dbs[shard].Set(b, buf.Bytes())
		atomic.AddUint64(&engine.numDocumentsStored, 1)
	}
}

func (engine *Engine) persistentStorageRemoveDocumentWorker(docId string, shard uint32) {
	// 得到key
	b := []byte(docId)

	// 从数据库删除该key
	engine.dbs[shard].Delete(b)
}

func (engine *Engine) persistentStorageInitWorker(shard int) {
	engine.dbs[shard].ForEach(func(k, v []byte) error {
		key, value := k, v
		// 得到docID
		docId := string(key)

		// 得到data
		buf := bytes.NewReader(value)
		dec := gob.NewDecoder(buf)
		var data types.DocumentIndexData
		err := dec.Decode(&data)
		if err == nil {
			// 添加索引
			engine.internalIndexDocument(docId, data, false)
		}
		return nil
	})
	engine.persistentStorageInitChannel <- true
}
