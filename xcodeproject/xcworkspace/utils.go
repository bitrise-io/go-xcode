package xcworkspace

import "path/filepath"

func IsWorkspace(pth string) bool {
	return filepath.Ext(pth) == XCWorkspaceExtension
}
