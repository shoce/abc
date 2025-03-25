package main
import ("crypto/sha256"; "io"; "fmt"; "os"; "github.com/btcsuite/btcutil/base58")
func main() {
	h:= sha256.New()
	h.Write([]byte{0x80})
	n, err := io.Copy(h, os.Stdin)
	if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
	fmt.Fprintln(os.Stderr, n, "bytes long data hashed")
	hs := h.Sum(nil)
	fmt.Println(base58.Encode(hs))
}
