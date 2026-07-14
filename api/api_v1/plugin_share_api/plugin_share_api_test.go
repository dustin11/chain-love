package plugin_share_api

import "testing"

func TestDetectBackgroundExtension(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{name: "webp", data: []byte("RIFF\x10\x00\x00\x00WEBPVP8 "), want: ".webp"},
		{name: "png", data: []byte("\x89PNG\r\n\x1a\n"), want: ".png"},
		{name: "jpeg", data: []byte("\xff\xd8\xff\xe0\x00\x10JFIF"), want: ".jpg"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := detectBackgroundExtension(test.data)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("extension = %q, want %q", got, test.want)
			}
		})
	}
	if _, err := detectBackgroundExtension([]byte("not an image")); err == nil {
		t.Fatal("expected unsupported background error")
	}
}
