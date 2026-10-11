#include "process-exit.h"

#ifdef _WIN32
#include <windows.h>
#else
#include <cerrno>
#include <unistd.h>
#ifdef __APPLE__
#include <sys/event.h>
#else
#include <poll.h>
#include <sys/syscall.h>
#endif
#endif

bool waitForProcessExit(int32_t pid)
{
	if (pid <= 0)
		return false;
#ifdef _WIN32
	HANDLE process = OpenProcess(SYNCHRONIZE, FALSE, pid);
	if (!process)
		return GetLastError() == ERROR_INVALID_PARAMETER;
	const DWORD result = WaitForSingleObject(process, INFINITE);
	CloseHandle(process);
	return result == WAIT_OBJECT_0;
#elif defined(__APPLE__)
	const int queue = kqueue();
	if (queue < 0)
		return false;
	struct kevent event;
	EV_SET(&event, pid, EVFILT_PROC, EV_ADD | EV_ONESHOT, NOTE_EXIT, 0,
	       nullptr);
	int result = kevent(queue, &event, 1, nullptr, 0, nullptr);
	if (result < 0) {
		const bool exited = errno == ESRCH;
		close(queue);
		return exited;
	}
	do {
		result = kevent(queue, nullptr, 0, &event, 1, nullptr);
	} while (result < 0 && errno == EINTR);
	close(queue);
	return result == 1 && !(event.flags & EV_ERROR);
#else
	const int process = syscall(SYS_pidfd_open, pid, 0);
	if (process < 0)
		return errno == ESRCH;
	struct pollfd event{process, POLLIN, 0};
	int result;
	do {
		result = poll(&event, 1, -1);
	} while (result < 0 && errno == EINTR);
	close(process);
	return result == 1 && (event.revents & POLLIN);
#endif
}
