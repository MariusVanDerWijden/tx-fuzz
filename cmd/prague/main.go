package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/MariusVanDerWijden/tx-fuzz/helper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	fmt.Println("Touching contracts")
	testTouchContracts()
	fmt.Println("2537")
	test2537()
	fmt.Println("2537")
	test2537Long()
	fmt.Println("3074")
	test3074()
	fmt.Println("7702")
	test7702()
	fmt.Println("2935")
	test2935()
	test7002()
	test7251()
}

func testTouchContracts() {
	// touch beacon root addr
	addresses := []common.Address{
		common.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02"), // beacon roots
		common.HexToAddress("0x09Fc772D0857550724b07B850a4323f39112aAaA"), // withdrawal requests
		common.HexToAddress("0x01aBEa29659e5e97C95107F20bb753cD3e09bBBb"), // consolidation requests
		common.HexToAddress("0xfffffffffffffffffffffffffffffffffffffffe"), // system address
		common.HexToAddress("0x25a219378dad9b3503c8268c9ca836a52427a4fb"), // history storage address
		common.HexToAddress("0x00000000219ab540356cbb839cbe05303d7705fa"), // mainnet deposit contract
		common.HexToAddress("0x4242424242424242424242424242424242424242"), // testnet deposit address
	}

	for _, addr := range addresses {
		helper.Exec(addr, []byte{}, false)                       // no data
		helper.Exec(addr, []byte{1}, false)                      // 1 byte of data
		helper.Exec(addr, crypto.Keccak256([]byte{1})[:], false) // 32 bytes of data
		helper.Exec(addr, make([]byte, 20), false)
		helper.Exec(addr, make([]byte, 2048), false) // 2048 bytes of data
	}
}

func test3074() {
	// auth
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0xf6, 0x80, 0x55}, 200000)
	helper.Execute([]byte{0x64, 0xff, 0xff, 0xff, 0xff, 0x64, 0xff, 0xff, 0xff, 0xff, 0x64, 0xff, 0xff, 0xff, 0xff, 0xf6, 0x80, 0x55}, 200000)
	// authcall
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0xf7, 0x80, 0x55}, 200000)
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5a, 0xf7, 0x80, 0x55}, 200000)
	helper.Execute([]byte{0x64, 0xff, 0xff, 0xff, 0xff, 0x5f, 0x64, 0xff, 0xff, 0xff, 0xff, 0x5f, 0x5f, 0x5f, 0x5a, 0xf7, 0x80, 0x55}, 200000)
	fmt.Println("Execution tests")
	vectors := [][]byte{
		common.FromHex("000f6617e03f2800b69a0b018d3062535ec761c6648a4c73be71f97885e28505f67f0d4ee582093960a99757587fe74e6ec173477c4ca05310e25158152ff99d4f0000000000000000000000000000000000000000000000000000000000000001"),
		common.FromHex("0x6001615dc06000f615600155"),
		common.FromHex("0x366000600037366000738a0a19589531694250d570040a0c4b74576919b8f6156001555a738a0a19589531694250d570040a0c4b74576919b83b505a03600003600255"),
		common.FromHex("0x0068bd8d0735aebb7ee746c81daaa23146eedc17e5882268d2be5b29f422bad1e67aa576acbef01fb78a701db206aa9d7de224f4cc5f5a3b4d0b9a9bb60f9aa5b60000000000000000000000000000000000000000000000000000000000000000"),
		common.FromHex("0x01f2f89f718c81bfdac9f08fc1cd6b91de657519f87ba41bcd6393a95833a55ef2616a0a70c31ec6801a9ab304dda6d14a67d0897f588b15bf73d99371b9db44fa0000000000000000000000000000000000000000000000000000000000000001aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
	}
	for _, vec := range vectors {
		helper.Execute(vec, 300000)
	}
}

func test2537() {
	vectors := [][]byte{
		{},
		{0x5f},             // small input
		make([]byte, 4096), // big input
		common.FromHex("000000000000000000000000000000000572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e00000000000000000000000000000000166a9d8cabc673a322fda673779d8e3822ba3ecb8670e461f73bb9021d5fd76a4c56d9d4cd16bd1bba86881979749d280000000000000000000000000000000009ece308f9d1f0131765212deca99697b112d61f9be9a5f1f3780a51335b3ff981747a0b2ca2179b96d2c0c9024e522400000000000000000000000000000000032b80d3a6f5b09f8a84623389c5f80ca69a0cddabc3097f9d9c27310fd43be6e745256c634af45ca3473b0590ae30d1"),
		common.FromHex("0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e10000000000000000000000000000000000000000000000000000000000000000"),
		common.FromHex("0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e10000000000000000000000000000000000000000000000000000000000000011"),
		common.FromHex("00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be"),
		common.FromHex("00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be0000000000000000000000000000000000000000000000000000000000000000"),
		common.FromHex("00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be0000000000000000000000000000000000000000000000000000000000000011"),
		common.FromHex("0000000000000000000000000000000014406e5bfb9209256a3820879a29ac2f62d6aca82324bf3ae2aa7d3c54792043bd8c791fccdb080c1a52dc68b8b69350"),
		common.FromHex("0000000000000000000000000000000014406e5bfb9209256a3820879a29ac2f62d6aca82324bf3ae2aa7d3c54792043bd8c791fccdb080c1a52dc68b8b69350000000000000000000000000000000000e885bb33996e12f07da69073e2c0cc880bc8eff26d2a724299eb12d54f4bcf26f4748bb020e80a7e3794a7b0e47a641"),
		common.FromHex("000000000000000000000000000000000572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e00000000000000000000000000000000166a9d8cabc673a322fda673779d8e3822ba3ecb8670e461f73bb9021d5fd76a4c56d9d4cd16bd1bba86881979749d2800000000000000000000000000000000122915c824a0857e2ee414a3dccb23ae691ae54329781315a0c75df1c04d6d7a50a030fc866f09d516020ef82324afae0000000000000000000000000000000009380275bbc8e5dcea7dc4dd7e0550ff2ac480905396eda55062650f8d251c96eb480673937cc6d9d6a44aaa56ca66dc000000000000000000000000000000000b21da7955969e61010c7a1abc1a6f0136961d1e3b20b1a7326ac738fef5c721479dfd948b52fdf2455e44813ecfd8920000000000000000000000000000000008f239ba329b3967fe48d718a36cfe5f62a7e42e0bf1c1ed714150a166bfbd6bcf6b3b58b975b9edea56d53f23a0e8490000000000000000000000000000000006e82f6da4520f85c5d27d8f329eccfa05944fd1096b20734c894966d12a9e2a9a9744529d7212d33883113a0cadb9090000000000000000000000000000000017d81038f7d60bee9110d9c0d6d1102fe2d998c957f28e31ec284cc04134df8e47e8f82ff3af2e60a6d9688a4563477c00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000d1b3cc2c7027888be51d9ef691d77bcb679afda66c73f17f9ee3837a55024f78c71363275a75d75d86bab79f74782aa0000000000000000000000000000000013fa4d4a0ad8b1ce186ed5061789213d993923066dddaf1040bc3ff59f825c78df74f2d75467e25e0f55f8a00fa030ed"),
		common.FromHex("0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb00000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a2100000000000000000000000000000000112b98340eee2777cc3c14163dea3ec97977ac3dc5c70da32e6e87578f44912e902ccef9efe28d4a78b8999dfbca942600000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a21"),
		common.FromHex("10000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be00000000000000000000000000000000103121a2ceaae586d240843a398967325f8eb5a93e8fea99b62b9f88d8556c80dd726a4b30e84a36eeabaf3592937f2700000000000000000000000000000000086b990f3da2aeac0a36143b7d7c824428215140db1bb859338764cb58458f081d92664f9053b50b3fbd2e4723121b68000000000000000000000000000000000f9e7ba9a86a8f7624aa2b42dcc8772e1af4ae115685e60abc2c9b90242167acef3d0be4050bf935eed7c3b6fc7ba77e000000000000000000000000000000000d22c3652d0dc6f0fc9316e14268477c2049ef772e852108d269d9c38dba1d4802e8dae479818184c08f9a569d878451"),
		common.FromHex("000000000000000000000000000000000123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef00000000000000000000000000000000193fb7cedb32b2c3adc06ec11a96bc0d661869316f5e4a577a9f7c179593987beb4fb2ee424dbb2f5dd891e228b46c4a0000000000000000000000000000000000000000000000000000000000000002"),
		common.FromHex("00000000000000000000000000000000197bfd0342bbc8bee2beced2f173e1a87be576379b343e93232d6cef98d84b1d696e5612ff283ce2cfdccb2cfb65fa0c00000000000000000000000000000000184e811f55e6f9d84d77d2f79102fd7ea7422f4759df5bf7f6331d550245e3f1bcf6a30e3b29110d85e0ca16f9f6ae7a000000000000000000000000000000000f10e1eb3c1e53d2ad9cf2d398b2dc22c5842fab0a74b174f691a7e914975da3564d835cd7d2982815b8ac57f507348f000000000000000000000000000000000767d1c453890f1b9110fda82f5815c27281aba3f026ee868e4176a0654feea41a96575e0c4d58a14dbfbcc05b5010b10000000000000000000000000000000000000000000000000000000000000002"),
		common.FromHex("000000000000000000000000000000000123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef00000000000000000000000000000000193fb7cedb32b2c3adc06ec11a96bc0d661869316f5e4a577a9f7c179593987beb4fb2ee424dbb2f5dd891e228b46c4a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000112b98340eee2777cc3c14163dea3ec97977ac3dc5c70da32e6e87578f44912e902ccef9efe28d4a78b8999dfbca942600000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a210000000000000000000000000000000000000000000000000000000000000002"),
		common.FromHex("00000000000000000000000000000000197bfd0342bbc8bee2beced2f173e1a87be576379b343e93232d6cef98d84b1d696e5612ff283ce2cfdccb2cfb65fa0c00000000000000000000000000000000184e811f55e6f9d84d77d2f79102fd7ea7422f4759df5bf7f6331d550245e3f1bcf6a30e3b29110d85e0ca16f9f6ae7a000000000000000000000000000000000f10e1eb3c1e53d2ad9cf2d398b2dc22c5842fab0a74b174f691a7e914975da3564d835cd7d2982815b8ac57f507348f000000000000000000000000000000000767d1c453890f1b9110fda82f5815c27281aba3f026ee868e4176a0654feea41a96575e0c4d58a14dbfbcc05b5010b1000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000103121a2ceaae586d240843a398967325f8eb5a93e8fea99b62b9f88d8556c80dd726a4b30e84a36eeabaf3592937f2700000000000000000000000000000000086b990f3da2aeac0a36143b7d7c824428215140db1bb859338764cb58458f081d92664f9053b50b3fbd2e4723121b68000000000000000000000000000000000f9e7ba9a86a8f7624aa2b42dcc8772e1af4ae115685e60abc2c9b90242167acef3d0be4050bf935eed7c3b6fc7ba77e000000000000000000000000000000000d22c3652d0dc6f0fc9316e14268477c2049ef772e852108d269d9c38dba1d4802e8dae479818184c08f9a569d8784510000000000000000000000000000000000000000000000000000000000000002"),
		common.FromHex("000000000000000000000000000000000123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef00000000000000000000000000000000193fb7cedb32b2c3adc06ec11a96bc0d661869316f5e4a577a9f7c179593987beb4fb2ee424dbb2f5dd891e228b46c4a00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e100000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000d1b3cc2c7027888be51d9ef691d77bcb679afda66c73f17f9ee3837a55024f78c71363275a75d75d86bab79f74782aa0000000000000000000000000000000013fa4d4a0ad8b1ce186ed5061789213d993923066dddaf1040bc3ff59f825c78df74f2d75467e25e0f55f8a00fa030ed"),
		common.FromHex("0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e100000000000000000000000000000000197bfd0342bbc8bee2beced2f173e1a87be576379b343e93232d6cef98d84b1d696e5612ff283ce2cfdccb2cfb65fa0c00000000000000000000000000000000184e811f55e6f9d84d77d2f79102fd7ea7422f4759df5bf7f6331d550245e3f1bcf6a30e3b29110d85e0ca16f9f6ae7a000000000000000000000000000000000f10e1eb3c1e53d2ad9cf2d398b2dc22c5842fab0a74b174f691a7e914975da3564d835cd7d2982815b8ac57f507348f000000000000000000000000000000000767d1c453890f1b9110fda82f5815c27281aba3f026ee868e4176a0654feea41a96575e0c4d58a14dbfbcc05b5010b10000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e100000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000d1b3cc2c7027888be51d9ef691d77bcb679afda66c73f17f9ee3837a55024f78c71363275a75d75d86bab79f74782aa0000000000000000000000000000000013fa4d4a0ad8b1ce186ed5061789213d993923066dddaf1040bc3ff59f825c78df74f2d75467e25e0f55f8a00fa030ed"),
	}
	for i := 0xa; i < 0x14; i++ {
		for _, vec := range vectors {
			testBLS(i, vec)
		}
	}
}

// test7002 creates withdrawal requests in the EIP-7002 queue.
func test7002() {
	fmt.Println("test7002")

	cl, _ := helper.GetRealBackend()
	backend := ethclient.NewClient(cl)

	contract := common.HexToAddress("0x09Fc772D0857550724b07B850a4323f39112aAaA")
	value := big.NewInt(1000000000)
	inputs := [][]byte{
		// input data is pubkey(48) || amount(8)
		common.FromHex("b917cfdc0d25b72d55cf94db328e1629b7f4fde2c30cdacf873b664416f76a0c7f7cc50c9f72a3cb84be88144cde91250000000000000d80"),
		common.FromHex("b9812f7d0b1f2f969b52bbb2d316b0c2fa7c9dba85c428c5e6c27766bcc4b0c6e874702ff1eb1c7024b08524a977160100000000000f423f"),
	}
	for i, data := range inputs {
		tx := makeTxWithValue(contract, value, data)
		if err := backend.SendTransaction(context.Background(), tx); err != nil {
			panic("SendTransaction: " + err.Error())
		}
		receipt, err := bind.WaitMined(context.Background(), backend, tx)
		if err != nil {
			panic("WaitMined: " + err.Error())
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			panic(fmt.Sprintf("test7002 tx %d reverted", i))
		}
	}
}

// test7251 creates consolidation requests in the EIP-7251 queue.
func test7251() {
	fmt.Println("test7251")
	contract := common.HexToAddress("0x01aBEa29659e5e97C95107F20bb753cD3e09bBBb")
	inputs := [][]byte{
		// input data is source_blskey(48) || target_blskey(48)
		common.FromHex("b917cfdc0d25b72d55cf94db328e1629b7f4fde2c30cdacf873b664416f76a0c7f7cc50c9f72a3cb84be88144cde9125b9812f7d0b1f2f969b52bbb2d316b0c2fa7c9dba85c428c5e6c27766bcc4b0c6e874702ff1eb1c7024b08524a9771601"),
	}
	for _, data := range inputs {
		helper.Exec(contract, data, false)
	}
}

func test2537Long() {
	multiexpG1 := common.FromHex("00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be")
	testLongBLS(0x0d, multiexpG1)
	multiexpG2 := common.FromHex("000000000000000000000000000000000572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e00000000000000000000000000000000166a9d8cabc673a322fda673779d8e3822ba3ecb8670e461f73bb9021d5fd76a4c56d9d4cd16bd1bba86881979749d2800000000000000000000000000000000122915c824a0857e2ee414a3dccb23ae691ae54329781315a0c75df1c04d6d7a50a030fc866f09d516020ef82324afae0000000000000000000000000000000009380275bbc8e5dcea7dc4dd7e0550ff2ac480905396eda55062650f8d251c96eb480673937cc6d9d6a44aaa56ca66dc000000000000000000000000000000000b21da7955969e61010c7a1abc1a6f0136961d1e3b20b1a7326ac738fef5c721479dfd948b52fdf2455e44813ecfd8920000000000000000000000000000000008f239ba329b3967fe48d718a36cfe5f62a7e42e0bf1c1ed714150a166bfbd6bcf6b3b58b975b9edea56d53f23a0e8490000000000000000000000000000000006e82f6da4520f85c5d27d8f329eccfa05944fd1096b20734c894966d12a9e2a9a9744529d7212d33883113a0cadb9090000000000000000000000000000000017d81038f7d60bee9110d9c0d6d1102fe2d998c957f28e31ec284cc04134df8e47e8f82ff3af2e60a6d9688a4563477c00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000d1b3cc2c7027888be51d9ef691d77bcb679afda66c73f17f9ee3837a55024f78c71363275a75d75d86bab79f74782aa0000000000000000000000000000000013fa4d4a0ad8b1ce186ed5061789213d993923066dddaf1040bc3ff59f825c78df74f2d75467e25e0f55f8a00fa030ed")
	testLongBLS(0x10, multiexpG2)
	pairing := common.FromHex("000000000000000000000000000000000572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e00000000000000000000000000000000166a9d8cabc673a322fda673779d8e3822ba3ecb8670e461f73bb9021d5fd76a4c56d9d4cd16bd1bba86881979749d2800000000000000000000000000000000122915c824a0857e2ee414a3dccb23ae691ae54329781315a0c75df1c04d6d7a50a030fc866f09d516020ef82324afae0000000000000000000000000000000009380275bbc8e5dcea7dc4dd7e0550ff2ac480905396eda55062650f8d251c96eb480673937cc6d9d6a44aaa56ca66dc000000000000000000000000000000000b21da7955969e61010c7a1abc1a6f0136961d1e3b20b1a7326ac738fef5c721479dfd948b52fdf2455e44813ecfd8920000000000000000000000000000000008f239ba329b3967fe48d718a36cfe5f62a7e42e0bf1c1ed714150a166bfbd6bcf6b3b58b975b9edea56d53f23a0e8490000000000000000000000000000000006e82f6da4520f85c5d27d8f329eccfa05944fd1096b20734c894966d12a9e2a9a9744529d7212d33883113a0cadb9090000000000000000000000000000000017d81038f7d60bee9110d9c0d6d1102fe2d998c957f28e31ec284cc04134df8e47e8f82ff3af2e60a6d9688a4563477c00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000d1b3cc2c7027888be51d9ef691d77bcb679afda66c73f17f9ee3837a55024f78c71363275a75d75d86bab79f74782aa0000000000000000000000000000000013fa4d4a0ad8b1ce186ed5061789213d993923066dddaf1040bc3ff59f825c78df74f2d75467e25e0f55f8a00fa030ed")
	testLongBLS(0x11, pairing)
}

func testLongBLS(prec int, input []byte) {
	fmt.Printf("Testing 0x%02x\n", prec)
	addr, err := deployPrecompileCaller(fmt.Sprintf("%02x", prec))
	if err != nil {
		panic(err)
	}
	base := input
	for i := 0; i < 20; i++ {
		helper.Exec(addr, base, false)
		base = append(base, base...)
	}
}

func testBLS(prec int, input []byte) {
	addr, err := deployPrecompileCaller(fmt.Sprintf("%02x", prec))
	if err != nil {
		panic(err)
	}
	helper.Exec(addr, input, false)
}

/*
pragma solidity >=0.7.0 <0.9.0;

	contract BlobCaller {
	    bool _ok;
	    bytes out;

	    fallback (bytes calldata _input) external returns (bytes memory _output) {
	        address precompile = address(0x0A);
	        (bool ok, bytes memory output) = precompile.call{gas: 500000}(_input);
	        _output = output;
	        // Store return values to trigger sstore
	        _ok = ok;
	        out = output;
	    }
	}
*/
func deployPrecompileCaller(precompile string) (common.Address, error) {
	bytecode1 := "6080604052348015600e575f80fd5b506104568061001c5f395ff3fe608060405234801561000f575f80fd5b505f3660605f60"
	bytecode2 := "90505f808273ffffffffffffffffffffffffffffffffffffffff1661c35087876040516100459291906100fe565b5f604051808303815f8787f1925050503d805f811461007f576040519150601f19603f3d011682016040523d82523d5f602084013e610084565b606091505b5091509150809350815f806101000a81548160ff02191690831515021790555080600190816100b39190610350565b50505050915050805190602001f35b5f81905092915050565b828183375f83830152505050565b5f6100e583856100c2565b93506100f28385846100cc565b82840190509392505050565b5f61010a8284866100da565b91508190509392505050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061019157607f821691505b6020821081036101a4576101a361014d565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026102067fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826101cb565b61021086836101cb565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61025461024f61024a84610228565b610231565b610228565b9050919050565b5f819050919050565b61026d8361023a565b6102816102798261025b565b8484546101d7565b825550505050565b5f90565b610295610289565b6102a0818484610264565b505050565b5b818110156102c3576102b85f8261028d565b6001810190506102a6565b5050565b601f821115610308576102d9816101aa565b6102e2846101bc565b810160208510156102f1578190505b6103056102fd856101bc565b8301826102a5565b50505b505050565b5f82821c905092915050565b5f6103285f198460080261030d565b1980831691505092915050565b5f6103408383610319565b9150826002028217905092915050565b61035982610116565b67ffffffffffffffff81111561037257610371610120565b5b61037c825461017a565b6103878282856102c7565b5f60209050601f8311600181146103b8575f84156103a6578287015190505b6103b08582610335565b865550610417565b601f1984166103c6866101aa565b5f5b828110156103ed578489015182556001820191506020850194506020810190506103c8565b8683101561040a5784890151610406601f891682610319565b8355505b6001600288020188555050505b50505050505056fea2646970667358221220bc28435cfa3208db8cae33e216a1ff54a6e5dce073695cad36274cc363055c5564736f6c63430008190033"
	// The byte in between bytecode1 and bytecode2 denotes the precompile which we want to call
	return helper.Deploy(fmt.Sprintf("%v%v%v", bytecode1, precompile, bytecode2))
}

// makeTxWithValue creates a transaction invoking addr, sending eth along with calldata.
func makeTxWithValue(addr common.Address, value *big.Int, data []byte) *types.Transaction {
	ctx := context.Background()
	cl, sk := helper.GetRealBackend()
	backend := ethclient.NewClient(cl)
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(ctx, sender)
	if err != nil {
		panic(err)
	}
	chainid, err := backend.ChainID(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	gp, err := backend.SuggestGasPrice(ctx)
	if err != nil {
		panic(err)
	}
	tip, err := backend.SuggestGasTipCap(ctx)
	if err != nil {
		panic(err)
	}
	return types.MustSignNewTx(sk, types.NewCancunSigner(chainid), &types.DynamicFeeTx{
		ChainID:   chainid,
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: gp,
		Gas:       200_000,
		Value:     value,
		To:        &addr,
		Data:      data,
	})
}

func test7702() {
	// authenticate self
	selfAddr := common.HexToAddress(txfuzz.ADDR)
	unsigned := &types.Authorization{
		ChainID: helper.ChainID().Uint64(),
		Address: selfAddr,
		Nonce:   helper.Nonce(selfAddr),
	}
	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	self, _ := types.SignAuth(unsigned, sk)
	helper.ExecAuth(selfAddr, []byte{}, &types.AuthorizationList{self})
	// authenticate self twice
	helper.ExecAuth(selfAddr, []byte{}, &types.AuthorizationList{self, self})
	// authenticate self twice with different nonces
	self2 := *self
	self2.Nonce = helper.Nonce(selfAddr) + 1
	self2P, _ := types.SignAuth(&self2, sk)
	helper.ExecAuth(selfAddr, []byte{}, &types.AuthorizationList{self, self2P})
	// unsigned authorization
	helper.ExecAuth(selfAddr, []byte{}, &types.AuthorizationList{unsigned})
	// many authorizations
	var list types.AuthorizationList
	for i := 0; i < 1024; i++ {
		list = append(list, self)
	}
	helper.ExecAuth(selfAddr, []byte{}, &list)
	// too many authorizations
	for i := 0; i < 1024*1023; i++ {
		list = append(list, self)
	}
	helper.ExecAuth(selfAddr, []byte{}, &list)
}

func test2935() {
	contr, err := deploy2935Caller()
	if err != nil {
		panic(err)
	}
	addresses := []common.Address{
		contr,
		params.HistoryStorageAddress,
	}

	cl, _ := helper.GetRealBackend()

	for _, addr := range addresses {
		// empty bytes
		helper.Exec(addr, []byte{}, false)
		// 32 bytes random
		var randomBytes [32]byte
		rand.Read(randomBytes[:])
		helper.Exec(addr, randomBytes[:], false)
		// 33 bytes
		var bigBytes [33]byte
		rand.Read(bigBytes[:])
		helper.Exec(addr, bigBytes[:], false)
		// 32 bytes 0
		var zeroBytes [32]byte
		helper.Exec(addr, zeroBytes[:], false)
		// 1
		helper.Exec(addr, []byte{1}, false)
		// block number specifics
		client := ethclient.NewClient(cl)
		currentBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			panic(err)
		}
		// current block number
		blocknumbers := []uint64{
			currentBlock,
			currentBlock + 1,
			currentBlock - 1,
			currentBlock - 8192,
			currentBlock - 256,
			currentBlock - 255,
		}
		for _, number := range blocknumbers {
			helper.Exec(addr, binary.BigEndian.AppendUint64([]byte{}, number), false)
		}
	}
}

/*
pragma solidity >=0.7.0 <0.9.0;

	contract EIP2935Caller {
	    bool _ok;
	    bytes out;
	    fallback (bytes calldata _input) external returns (bytes memory _output) {
	        address contrAddr = address(0x0AAE40965E6800cD9b1f4b05ff21581047E3F91e);
	        (bool ok, bytes memory output) = contrAddr.call{gas: 500000}(_input);
	        _output = output;
	        // Store return values to trigger sstore
		    _ok = ok;
		    out = output;
	    }
	}
*/
func deploy2935Caller() (common.Address, error) {
	bytecode1 := "6080604052348015600e575f80fd5b506104698061001c5f395ff3fe608060405234801561000f575f80fd5b505f3660605f730aae40965e6800cd9b1f4b05ff21581047e3f91e90505f808273ffffffffffffffffffffffffffffffffffffffff166207a1208787604051610059929190610112565b5f604051808303815f8787f1925050503d805f8114610093576040519150601f19603f3d011682016040523d82523d5f602084013e610098565b606091505b5091509150809350815f806101000a81548160ff02191690831515021790555080600190816100c79190610364565b50505050915050805190602001f35b5f81905092915050565b828183375f83830152505050565b5f6100f983856100d6565b93506101068385846100e0565b82840190509392505050565b5f61011e8284866100ee565b91508190509392505050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806101a557607f821691505b6020821081036101b8576101b7610161565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f6008830261021a7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826101df565b61022486836101df565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61026861026361025e8461023c565b610245565b61023c565b9050919050565b5f819050919050565b6102818361024e565b61029561028d8261026f565b8484546101eb565b825550505050565b5f90565b6102a961029d565b6102b4818484610278565b505050565b5b818110156102d7576102cc5f826102a1565b6001810190506102ba565b5050565b601f82111561031c576102ed816101be565b6102f6846101d0565b81016020851015610305578190505b610319610311856101d0565b8301826102b9565b50505b505050565b5f82821c905092915050565b5f61033c5f1984600802610321565b1980831691505092915050565b5f610354838361032d565b9150826002028217905092915050565b61036d8261012a565b67ffffffffffffffff81111561038657610385610134565b5b610390825461018e565b61039b8282856102db565b5f60209050601f8311600181146103cc575f84156103ba578287015190505b6103c48582610349565b86555061042b565b601f1984166103da866101be565b5f5b82811015610401578489015182556001820191506020850194506020810190506103dc565b8683101561041e578489015161041a601f89168261032d565b8355505b6001600288020188555050505b50505050505056fea264697066735822122033feaed59f5038b726e0ad09706fc2b959f0781cd00f5b4961c7ce6a676215f764736f6c634300081a0033"
	return helper.Deploy(bytecode1)
}
