package flock

import (
	"os"
)

// Lock_EX_NB 排它锁 -> 非阻塞
// 排它锁，也叫写锁。某个进程首次获取排他锁后，会生成一个锁类型的变量L，类型标记为排他锁。
// 其它进程获取任何类型的锁的时候，都会获取失败。
func Lock_EX_NB(f *os.File) {
}

// Lock_EX 排它锁 -> 阻塞
func Lock_EX(f *os.File) {
}

// Lock_SH_NB 共享锁 -> 非阻塞
// 共享锁，也叫读锁。某个进程首次获取共享锁后，会生成一个锁类型的变量L，类型标记为共享锁。
// 其它进程获取读锁的时候，L中的计数器加1，表示又有一个进程获取到了共享锁。这个时候如果有进程来获取排它锁，会获取失败。
func Lock_SH_NB(f *os.File) {
}

// Lock_SH 共享锁 -> 阻塞
func Lock_SH(f *os.File) {
}

// Lock_UN 解锁
func Lock_UN(f *os.File) {
}
