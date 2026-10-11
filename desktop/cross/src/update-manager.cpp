#include "update-manager.h"
#include "process-exit.h"

#include <cstdio>

#ifdef _WIN32
#include <process.h>
#include <windows.h>
#else
#include <signal.h>
#include <spawn.h>
#include <sys/stat.h>
#include <unistd.h>
extern char **environ;
#endif

UpdateManager::UpdateManager(const std::string &current_path,
			     const std::string &staged_path, int32_t pid)
	: current_path_(current_path),
	  staged_path_(staged_path),
	  pid_(pid)
{
}

bool UpdateManager::execute()
{
	if (!waitForProcessExit(pid_))
		return false;
	if (!SwapBinary())
		return false;
	if (!Relaunch()) {
		// Retain the failed candidate and restore the last runnable
		// installation.
		if (rename(current_path_.c_str(), staged_path_.c_str()) != 0 ||
		    rename((current_path_ + ".old").c_str(),
			   current_path_.c_str()) != 0) {
			fprintf(stderr, "error: failed to restore previous "
					"installation\n");
		}
		return false;
	}
	remove((current_path_ + ".old").c_str());
	return true;
}

bool UpdateManager::SwapBinary()
{
	std::string backup_path = current_path_ + ".old";

	// Remove old backup if exists.
	remove(backup_path.c_str());

	// Rename current to backup.
	if (rename(current_path_.c_str(), backup_path.c_str()) != 0) {
		fprintf(stderr, "error: failed to rename current to backup\n");
		return false;
	}

	// Move staged into place.
	if (rename(staged_path_.c_str(), current_path_.c_str()) != 0) {
		// Rollback: move backup back.
		rename(backup_path.c_str(), current_path_.c_str());
		fprintf(stderr, "error: failed to move staged into place\n");
		return false;
	}

	// Keep the backup until the replacement has successfully launched.
	return true;
}

bool UpdateManager::Relaunch()
{
#ifdef _WIN32
	// Spawn new process and exit.
	STARTUPINFOA si;
	PROCESS_INFORMATION pi;
	memset(&si, 0, sizeof(si));
	si.cb = sizeof(si);
	memset(&pi, 0, sizeof(pi));

	if (!CreateProcessA(current_path_.c_str(), nullptr, nullptr, nullptr,
			    FALSE, 0, nullptr, nullptr, &si, &pi)) {
		fprintf(stderr, "error: failed to Relaunch\n");
		return false;
	}
	CloseHandle(pi.hProcess);
	CloseHandle(pi.hThread);
	return true;
#else
	// posix_spawn reports exec errors to the updater, unlike a successful
	// fork.
	if (chmod(current_path_.c_str(), 0755) != 0)
		return false;
	pid_t child;
	std::string executable = current_path_;
	char *args[] = {executable.data(), nullptr};
	const int result = posix_spawn(&child, current_path_.c_str(), nullptr,
				       nullptr, args, environ);
	if (result != 0) {
		fprintf(stderr, "error: failed to Relaunch (%d)\n", result);
		return false;
	}
	return true;
#endif
}
