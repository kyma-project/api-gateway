package v2

import "regexp"

const istioValidPathRegex = `^((\/([A-Za-z0-9-._~!$&'()+,;=:@]|%[0-9a-fA-F]{2})*)|(\/\{\*{1,2}\}))+$|^\/\*$`

// isConvertiblePath checks if the path is convertible to Istio VirtualService path compatible format
// this regex allows one exception: /* which is translated in module to be equal /{**}
func isConvertiblePath(path string) bool {
	validPathRegex := regexp.MustCompile(istioValidPathRegex)
	return validPathRegex.MatchString(path)
}
