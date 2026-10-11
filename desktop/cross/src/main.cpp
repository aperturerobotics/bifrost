#include "loading-window.h"
#include "pipe-client.h"
#include "process-exit.h"
#include "update-manager.h"
#include <cstdio>
#include <cstring>
#include <filesystem>
#include <string>
#include <vector>

// spacewave-helper: native helper for loading screen, self-update and
// uninstall.
// Modes:
//   --loading --icon <path> --pipe-root <dir> --pipe-id <uuid>
//   --update --current <path> --staged <path> --pid <int> --pipe-root <dir>
//   --pipe-id <uuid>
//   --remove <path> [--remove <path>...] --pid <int>

struct Args {
	bool loading = false;
	bool update = false;
	std::string iconPath;
	std::string pipeRoot;
	std::string pipeId;
	std::string currentPath;
	std::string stagedPath;
	std::vector<std::string> removePaths;
	int pid = 0;
};

static Args parseArgs(int argc, char *argv[])
{
	Args args;
	for (int i = 1; i < argc; i++) {
		if (strcmp(argv[i], "--loading") == 0) {
			args.loading = true;
		} else if (strcmp(argv[i], "--update") == 0) {
			args.update = true;
		} else if (strcmp(argv[i], "--icon") == 0 && i + 1 < argc) {
			args.iconPath = argv[++i];
		} else if (strcmp(argv[i], "--pipe-root") == 0 &&
			   i + 1 < argc) {
			args.pipeRoot = argv[++i];
		} else if (strcmp(argv[i], "--pipe-id") == 0 && i + 1 < argc) {
			args.pipeId = argv[++i];
		} else if (strcmp(argv[i], "--current") == 0 && i + 1 < argc) {
			args.currentPath = argv[++i];
		} else if (strcmp(argv[i], "--staged") == 0 && i + 1 < argc) {
			args.stagedPath = argv[++i];
		} else if (strcmp(argv[i], "--remove") == 0 && i + 1 < argc) {
			args.removePaths.push_back(argv[++i]);
		} else if (strcmp(argv[i], "--pid") == 0 && i + 1 < argc) {
			args.pid = atoi(argv[++i]);
		}
	}
	return args;
}

// removeAfterExit deletes each path once process pid has exited, so the
// uninstalling launcher no longer holds its own executable.
static int removeAfterExit(const std::vector<std::string> &paths, int pid)
{
	if (!waitForProcessExit(int32_t(pid))) {
		fprintf(stderr, "error: --remove requires an exiting --pid\n");
		return 1;
	}
	int status = 0;
	for (const std::string &path : paths) {
		std::error_code err;
		std::filesystem::remove_all(path, err);
		if (err) {
			fprintf(stderr, "error: remove %s: %s\n", path.c_str(),
				err.message().c_str());
			status = 1;
		}
	}
	return status;
}

int main(int argc, char *argv[])
{
	Args args = parseArgs(argc, argv);

	// Removal runs after the uninstalling process exits, so it has no pipe.
	if (!args.removePaths.empty())
		return removeAfterExit(args.removePaths, args.pid);

	if (args.pipeRoot.empty() || args.pipeId.empty()) {
		fprintf(stderr, "error: --pipe-root and --pipe-id required\n");
		return 1;
	}

#ifdef _WIN32
	std::string pipePath = "\\\\.\\pipe\\aptre\\" + args.pipeId;
#else
	std::string pipePath = args.pipeRoot + "/.pipe-" + args.pipeId;
#endif
	PipeClient pipe(pipePath);
	if (!pipe.connect()) {
		fprintf(stderr, "error: failed to connect to pipe at %s\n",
			pipePath.c_str());
		return 1;
	}

	// Send ready event.
	HelperEvent readyEvt{EventType::Ready};
	if (!pipe.sendEvent(readyEvt)) {
		fprintf(stderr, "error: failed to send ready event\n");
		return 1;
	}

	if (args.loading) {
		LoadingWindow *window = createLoadingWindow();
		if (!window->create("Spacewave", args.iconPath)) {
			fprintf(stderr, "error: failed to create window\n");
			return 1;
		}

		// Wire pipe messages to window.
		pipe.startReadLoop([window](const HelperMessage &msg) {
			window->handleMessage(msg);
		});

		// Wire window events back to pipe.
		window->onRetry = [&pipe]() {
			pipe.sendEvent(HelperEvent{EventType::Retry});
		};
		window->onCancel = [&pipe]() {
			pipe.sendEvent(HelperEvent{EventType::Cancel});
			pipe.close();
		};

		window->show();
		window->runEventLoop();
		delete window;

	} else if (args.update) {
		if (args.currentPath.empty() || args.stagedPath.empty() ||
		    args.pid <= 0) {
			fprintf(stderr, "error: --update requires --current, "
					"--staged, --pid\n");
			return 1;
		}

		UpdateManager updater(args.currentPath, args.stagedPath,
				      int32_t(args.pid));
		if (!updater.execute()) {
			fprintf(stderr, "error: update failed\n");
			return 1;
		}

	} else {
		fprintf(stderr,
			"error: specify --loading, --update or --remove\n");
		return 1;
	}

	pipe.close();
	return 0;
}
