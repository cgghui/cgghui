package flock

import (
	"os"
	"syscall"
)

// Lock_EX_NB 排它锁 -> 非阻塞
// 排它锁，也叫写锁。某个进程首次获取排他锁后，会生成一个锁类型的变量L，类型标记为排他锁。
// 其它进程获取任何类型的锁的时候，都会获取失败。
func Lock_EX_NB(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

// Lock_EX 排它锁 -> 阻塞
func Lock_EX(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// Lock_SH_NB 共享锁 -> 非阻塞
// 共享锁，也叫读锁。某个进程首次获取共享锁后，会生成一个锁类型的变量L，类型标记为共享锁。
// 其它进程获取读锁的时候，L中的计数器加1，表示又有一个进程获取到了共享锁。这个时候如果有进程来获取排它锁，会获取失败。
func Lock_SH_NB(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
}

// Lock_SH 共享锁 -> 阻塞
func Lock_SH(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
}

// Lock_UN 解锁
func Lock_UN(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

// 阻塞
//
// 阻塞的意思是说，新的进程发现当前的文件（数据）被加锁后，会一直处于等待状态，直到锁被释放，才会继续下一步的行为。
//
// 非阻塞
//
// 非阻塞的意思是说，新的进程发现当前的文件（数据）被加锁后，立即返回异常。业务上需要根据具体的业务场景对该异常进行处理。
//
// 阻塞和非阻塞其实是进程遇到锁的时候的两种处理模式。
