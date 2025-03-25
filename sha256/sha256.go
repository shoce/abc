package main
import ("crypto/sha256"; "io"; "fmt"; "os"; "encoding/hex")
func main() {
	h:= sha256.New()
	n, err := io.Copy(h, os.Stdin)
	if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
	fmt.Fprintln(os.Stderr, n, "bytes long data hashed")
	hs := h.Sum(nil)
	fmt.Println(hex.EncodeToString(hs))
}
