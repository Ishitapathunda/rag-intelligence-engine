package chunker

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"
)

type Chunk struct {
	Id       string `json:"id"`
	Source   string `json:"source"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

func ChunkText(source, text string, chunkSize, overlap int) []Chunk {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	stride := chunkSize - overlap
	var chunks []Chunk
	pos := 0

	for i := 0; i < len(words); i += stride {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := Chunk{
			Id:       makeID(source, pos),
			Source:   source,
			Text:     strings.Join(words[i:end], " "),
			Position: pos,
		}

		chunks = append(chunks, chunk)
		pos++

		if end == len(words) {
			break
		}
	}

	return chunks
}

func makeID(source string, pos int) string {
	raw := source + "-" + strconv.Itoa(pos)
	h := sha1.Sum([]byte(raw))
	return hex.EncodeToString(h[:])[:16]
}
