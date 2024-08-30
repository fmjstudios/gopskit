package util

func DeepMergeMap(dst, src map[string]interface{}) error {
	for srcK, srcV := range src {

		// check if interface is another map
		if srcMap, ok := srcV.(map[string]interface{}); ok {
			if dstVal, ok := dst[srcK]; ok {
				if dstMap, ok := dstVal.(map[string]interface{}); ok {
					err := DeepMergeMap(dstMap, srcMap)
					if err != nil {
						return err
					}
					continue
				}
			} else {
				dst[srcK] = make(map[string]interface{})
			}
			err := DeepMergeMap(dst[srcK].(map[string]interface{}), srcMap)
			if err != nil {
				return err
			}
			// check if interface is a slice
		} else if srcSl, ok := srcV.([]interface{}); ok {
			if dstV, ok := dst[srcK]; ok {
				if dstSl, ok := dstV.([]interface{}); ok {
					dst[srcK] = append(dstSl, srcSl)
					continue
				}
			}
			dst[srcK] = srcSl
			// handle primitives
		} else {
			dst[srcK] = srcV
		}
	}
	return nil
}
