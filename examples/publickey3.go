// https://l1z2g9.github.io/2016/11/04/RSA-Encrypt-Decrypt-with-Golang/
package main
import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
)
/*
show by command prompt
openssl genrsa -out key.pem
openssl rsa -in key.pem  -pubout > key-pub.pem
echo polaris@studygolang.com | openssl rsautl \
     -encrypt \
     -pubin -inkey key-pub.pem \
 > cipher.txt
cat cipher.txt | openssl rsautl \
    -decrypt \
    -inkey key.pem
** OR encoding by base64 **
echo polaris@studygolang.com | openssl rsautl \ 
      -encrypt -pubin -inkey key-pub.pem \
      | openssl base64
openssl base64 -d | openssl rsautl -decrypt -inkey key.pem
*/
var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwOJK1RJBUwRu/5aCyktTaietXFMOAAkElhSq1M6BocVWs7yD
y592CX30Bl0Ul4faWM9EZSlhak8Ay1CdMNis+lBZanKmAO2bPmSIIYBDQE2BzLIo
MoJWi/Cd5PevioKSRPytqVB/S4+xz1IOD8Y407SZM3LfZ5XMfqC+VHpcnAycQ8iT
FK0s3yjImathFNF3U7fiEzU4G7PJRn8e9ntubDd1pXYABqrVF/REcd/3Rs/qrlhG
v3b7tAXZb2lkSLdCq3Md+BMksxUCoH3rZijSphbZSCdIrzofg+IG0y5WtdsBz6uw
Ol2QX/EUoEdO+xhLgaOFykUoWz037ZzkLEhKkQIDAQABAoIBAB+1lAPPSnnxYqYW
Ak5rb70l5LQm20haMyzRHPx7Loh/vq8xsKELCAardDCPoNEAfn7XJDFVSjSF5GWI
TS84j8de6jQ7wNqqNTleoZqQUX4Cv/H83+rdzoiW9/4qUet9Z7p7p7kMCMFNUDf7
D2C8f58eM4lnux52W/X9SwzsSMlGaGHcAKPz4vXUFWyt3naVtANhdkHjgKxA0Ev4
W7yRgpbOKruPKzBNTRXAqq+yHZj/pONtXl8do+plwhHU8CW0BPyvkU4DH52lxWza
mM71ow8UJC30FXF/NZ+wthFnRZO3/dhaeuNYgX7yAs3DhNn7Q8nzU4ujd8ug2OGf
iJ9C8YECgYEA32KthV7VTQRq3VuXMoVrYjjGf4+z6BVNpTsJAa4kF+vtTXTLgb4i
jpUrq6zPWZkQ/nR7+CuEQRUKbky4SSHTnrQ4yIWZTCPDAveXbLwzvNA1xD4w4nOc
JgG/WYiDtAf05TwC8p/BslX20Ox8ZAXUq6pkAeb1t8M2s7uDpZNtBMkCgYEA3QuU
vrrqYuD9bQGl20qJI6svH875uDLYFcUEu/vA/7gDycVRChrmVe4wU5HFErLNFkHi
uifiHo75mgBzwYKyiLgO5ik8E5BJCgEyA9SfEgRHjozIpnHfGbTtpfh4MQf2hFsy
DJbeeRFzQs4EW2gS964FK53zsEtnr7bphtvfY4kCgYEAgf6wr95iDnG1pp94O2Q8
+2nCydTcgwBysObL9Phb9LfM3rhK/XOiNItGYJ8uAxv6MbmjsuXQDvepnEp1K8nN
lpuWN8rXTOG6yG1A53wWN5iK0WrHk+BnTA7URcwVqJzAvO3RYVPqqlcwTKByOtrR
yhxcGmdHMusdWDaVA7PpS1ECgYATCGs/XQLPjsXje+/XCPzz+Epvd7fi12XpwfQd
Z5j/q82PsxC+SQCqR38bwwJwELs9/mBSXRrIPNFbJEzTTbinswl9YfGNUbAoT2AK
GmWz/HBY4uBoDIgEQ6Lu1o0q05+zV9LgaKExVYJSL0EKydRQRUimr8wK0wNTivFi
rk322QKBgHD3aEN39rlUesTPX8OAbPD77PcKxoATwpPVrlH8YV7TxRQjs5yxLrxL
S21UkPRxuDS5CMXZ+7gA3HqEQTXanNKJuQlsCIWsvipLn03DK40nYj54OjEKYo/F
UgBgrck6Zhxbps5leuf9dhiBrFUPjC/gcfyHd/PYxoypHuQ3JUsJ
-----END RSA PRIVATE KEY-----
`)
var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwOJK1RJBUwRu/5aCyktT
aietXFMOAAkElhSq1M6BocVWs7yDy592CX30Bl0Ul4faWM9EZSlhak8Ay1CdMNis
+lBZanKmAO2bPmSIIYBDQE2BzLIoMoJWi/Cd5PevioKSRPytqVB/S4+xz1IOD8Y4
07SZM3LfZ5XMfqC+VHpcnAycQ8iTFK0s3yjImathFNF3U7fiEzU4G7PJRn8e9ntu
bDd1pXYABqrVF/REcd/3Rs/qrlhGv3b7tAXZb2lkSLdCq3Md+BMksxUCoH3rZijS
phbZSCdIrzofg+IG0y5WtdsBz6uwOl2QX/EUoEdO+xhLgaOFykUoWz037ZzkLEhK
kQIDAQAB
-----END PUBLIC KEY-----
`)
func RsaEncryptByPreSetPubKey() string {
	n, _ := new(big.Int).SetString("24349343452348953201209477858721354875245881458202672294652984377378513954748002477250933828219774703952578332297494223229725595176463711802920124930360492553186030821158773846902662847263120685557322462156596316871394035160273640449724455863863094140814233064652945361596472111169159061323006507670749392076044355771083774400487999226532334510138900864338047649454583762051951010712101235391104817996664455285600818344773697074965056427233256586264138950003914735074112527568699379597208762648078763602593269860453947862814755877433560650621539845829407336712267915875159364773551462882284084578152070138814976772753", 10)
	e, _ := strconv.ParseInt("10001", 16, 0)
	fmt.Printf("##RsaEncrypt2 n %x\n", n)
	fmt.Printf("##RsaEncrypt2 e %x\n", e)
	pubKey := rsa.PublicKey{n, int(e)}
	data, _ := rsa.EncryptPKCS1v15(rand.Reader, &pubKey, []byte("it's great for rsa"))
	return hex.EncodeToString(data)
}
func main() {
	msg := "polaris@studygolang.com"
	data, err := RsaEncrypt([]byte(msg))
	fmt.Printf("PKCS1v15 encrypted [%s] to \n[%x]\n", string(msg), data)
	ioutil.WriteFile("encrypted.txt", data, 0644)
	if err != nil {
		panic(err)
	}
	origData, err := RsaDecrypt(data)
	if err != nil {
		panic(err)
	}
	fmt.Println("origData >> ", string(origData))
	//cipherText, _ := hex.DecodeString("b6ee3caf14430003a20625ba1ea9ad31560ad203f7ecee46dd8e31f2dc47d278f3248bc0180e03571fdbf34a60aad7310468e6d6013fcfd6b785d1562411b44e089281adcc275a2037db3dec8b447b91162c859ab97372081c1bcb22a1fb33b1f72a06a54b1784d9f733aa1e869c6d64d45a7a78534714a773920ef7219b31f89092fc54f87ff371aeae5c3e59cdaad3fa05c24e781e06fcd46b35127a431bd85f62bafded95e3d31127159a0b5d13b77f11ecef94a037ac1d2f2c32fc0e6623cfe056127457f8f82631c33139a50fcd16c17e577b12f853cd55ffb16e099097dd76a21d987c536ac102b470e36881fc86f1667b505120a531458a116ca285b7")
	cipherText, _ := hex.DecodeString(RsaEncryptByPreSetPubKey())
	origData2, err := RsaDecrypt(cipherText)
	if err != nil {
		panic(err)
	}
	fmt.Println("origData2 >> ", string(origData2))
}
// 加密
func RsaEncrypt(origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	fmt.Println("Modulus : ", pub.N.String())
	fmt.Println(">>> ", pub.N)
	fmt.Printf("Modulus(Hex) : %X\n", pub.N)
	fmt.Println("Public Exponent : ", pub.E)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}
// 解密
func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}
