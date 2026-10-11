#pragma once
#include <cstdint>

// waitForProcessExit waits for an operating-system process-exit event. It
// returns true once process pid has exited, including when it already had.
bool waitForProcessExit(int32_t pid);
