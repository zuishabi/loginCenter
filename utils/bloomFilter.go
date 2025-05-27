package utils

import (
	"hash/crc32"
	"hash/fnv"
	"math"
	"sync"
)

// InitBF 初始化布隆过滤器
func InitBF(size float64) *BloomFilter {
	n := -size * math.Log(0.001) / (math.Ln2 * math.Ln2)
	s := (n / size) * math.Ln2
	data := make([]byte, int((n+7)/8))
	b := BloomFilter{data: data, hashSize: int(s), size: int(n)}
	return &b
}

type BloomFilter struct {
	data     []byte
	size     int
	hashSize int
	sync.Mutex
}

// SetItem 在布隆过滤器设置一个值
func (b *BloomFilter) SetItem(item string) {
	b.Lock()
	defer b.Unlock()
	for i := 0; i < b.hashSize; i++ {
		h := (hash1(item) + i*hash2(item)) % b.size
		if h < 0 {
			h += b.size
		}
		setBit(b.data, h)
	}
}

// GetItem 检查一个值是否在布隆过滤器内
func (b *BloomFilter) GetItem(item string) bool {
	b.Lock()
	defer b.Unlock()
	for i := 0; i < b.hashSize; i++ {
		h := (hash1(item) + i*hash2(item)) % b.size
		if h < 0 {
			h += b.size
		}
		if !getBit(b.data, h) {
			return false
		}
	}
	return true
}

func setBit(data []byte, n int) {
	data[n/8] |= 1 << (n % 8)
}

func getBit(data []byte, n int) bool {
	return (data[n/8] & (1 << (n % 8))) != 0
}

func hash1(s string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32())
}

func hash2(s string) int {
	return int(crc32.ChecksumIEEE([]byte(s)))
}
