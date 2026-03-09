package xcresult3

import "github.com/bitrise-io/go-utils/log"

func sendRemoteWarning(tag string, format string, v ...interface{}) {
	data := map[string]interface{}{}
	data["source"] = "deploy-to-bitrise-io"

	log.RWarnf("deploy-to-bitrise-io", tag, data, format, v...)
}
