package codegen

import "bytes"

type payloadBlob struct {
	name  string
	bytes []byte
}

func renderPayloadFile(pkg string, table *StringTable, blobs ...payloadBlob) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("package ")
	b.WriteString(pkg)
	b.WriteString("\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range blobs {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}
