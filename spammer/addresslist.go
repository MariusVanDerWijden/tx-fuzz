package spammer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	staticKeys = []string{
		"0xaf5ead4413ff4b78bc94191a2926ae9ccbec86ce099d65aaf469e9eb1a0fa87f",
		"0xe63135ee5310c0b34c551e4683ad926dce90062b15e43275f9189b0f29bc784c",
		"0xc216a7b5048e6ea2437b20bc9c7f9a57cc8aefd5aaf6a991c4db407218ed9e77",
		"0xc29d916f5b6ddd0aa2c827ab7333e40e91fda9ca980332b3c60cae5b7263dae7",
		"0xde1013a3fcdf8b204f90692478254e78126f9763f475cdaba9fc6a1abfb97db3",
		"0xb56fd1fd33b71c508e92bf71fae6e92c76fcd5c0df37a39ff3caa244ddba3c0f",
		"0x10ee939f126a0c5fc3c3cc0241cd945aa88f57eef36bde8707db51ceecfd9134",
		"0x03210fac527544d5b49b89763121b476c4ab66908b345916d6ad2c7740f23803",
		"0x8a6ccbba94844d3951c1e5581df9f8f87de8106b995a3a664d9130a2b72a4b96",
		"0x831a53a7994ac56f6c5d99d6371d7a8686f385995da2592aac82dda8b008b454",
		"0x5a1f678991d52ca86f2b0403727d19887a261199fd88908bde464f7bf13aa50b",
		"0x9823c88ce58b714d576741ae3b97c7f8445209e73596d7588bc5101171e863f4",
		"0x0bd4b408eb720ecc526cf72c88180a60102dd7fd049a5ad04e367c0d9bbc843e",
		"0x4e5ae855d26b0fdc601c8749ad578ba1ddd714997c40e0483479f0f9ff61c51d",
		"0xa0a0419689b82172eb4b9ee2670fef5e1bfa22909b51b7e4f9ab8a9719045a43",
		"0x7be763c508ccfd597ab4854614c55162506d454cd41627c46b264483bb8d2138",
		"0x5d6f77078b8f61d0258737a29c9d8fe347d70af006c109c5df91ae61a85be52b",
		"0x6737e36a661e8de28b5c448063a936bd6c310eab319338f4c503b8f18e935007",
		"0xbd5d6497a71ea419b0d0dc538016c238b8a6a630f6cdda299fcc4ce85858f95b",
		"0x8848fc11b20202e9f702ea05eed08e509eb8d5e32e61454f114bf2085392df75",
		"0x2b7aad673e69b9e4b2d7963ef27e503999f0bd5ff3e63f292e188d2e7df2fe60",
		"0xd2254db4e9d74fd2ab7c296f6bc06ab6beb11d8e3f685a582efc2b5e2cc8f86c",
		"0x5477ebb68a387dc9d4cf2c2b194bed9a027e7f817bd2cac553aca9fe1ec923ad",
		"0xb68a0d9d69df9697ce2c8e4e3a629e7400ddb88a93879a48788c8e8404b2ff90",
		"0xd2db07c60da1bf2048b84c1e09fe4d5bb1b6d0b0eb06bef801e1c2bac1c93d76",
		"0x9e759f9762cb967f96fe437cfa432e2889b2f972ca9f382756efb4998188be12",
		"0x18b886f1e77682ae7a92e9d1c29c13acfb2f493a69723b156510711526654e4f",
		"0x907c2e461495607062e0a7ad8bdb29d7129209ba1accbb478dbd3dee7671a1c8",
		"0xa63e812d650015a9ec0fc94c09b02cc9425e3e197be4a41f1b44a869dd3adace",
		"0xc9a3d46fa54409b795df80a363f43ae31e0c7d15f7d4c5062ce88ae7b78124b8",
		"0x6d41eb903d4f5b21e29a8d8558be7dce002e1b23298d2df7cee4dfdaff5b5980",
		"0xe969e9be3a7e87dc29699f61d5566c79d95803779575df98d054f4207b363333",
		"0xbf82a18972fadc7bf60d8c5bcbd37c8a55fe0cfbb17106e0617ebe6999b2bb61",
		"0x011b3a4adb79e6b372972d5a66ea1acdc44b1ca6ea9985af90fda4a12622926b",
		"0x43034269f49a0963cd45473c22d18693ebbd924f515a49b6fd9190dc96ec60de",
		"0x2a36981f40b25474da25277593836451c8ddc0a4fdd9131cd82dcef003a1c4ce",
		"0x59e57a2b3739c119d94e4e2ecf5c0f8430241e59d27f386f17d050d30f1d5d99",
		"0x8d6eb80f206eec85c773585295f850e159a27ba148360a34bb3355caab17f1b2",
		"0x1491bd992bff53671dd787070fdd54122c395690e248dc4eb32b0fb942a17cc2",
		"0xb7f637ebf0faef160984b039994b75fbec1714128eb1feafe92ce3f7e54bbbae",
		"0xfd0a5e72904e4c1497e315896fe6918c513f76503377005b70cf30dbd705fc50",
		"0x760f74cbbc74c4cdc521abcdbe8ae519091311a1cd3dfd04559848bf94a4af71",
		"0x4bc76ef2f36f8988d1c52eec26be1d31f212781bf918e57406a5e8ad14262c36",
		"0x716cce68bb7dace09415047aff1cf90af99ac6b81c128eb6c80ad9b739c3fc47",
		"0xba9d2c4fed88860a2fb2103590337e71e6720d4323e64dc24f1d9f2f98023c28",
		"0xc9eb9929bec8348030b917fca6e0c4a4b08059c831e9c40250eddce6a942d61f",
		"0xcbd0db2dabc113c02254bc50efd457bcc446338cde9c7f5c93be4f155541bf8e",
		"0x0a6c4e2207afd4c74d78883c4ef8c526dd164b67faf8c8da928d4a7c694fa49c",
		"0xde664e3cde706b7dfc33e1c561e9c63f6223c68668507a04804644aa7f56c8ce",
		"0xbefc32493440df7b3a3951e824570f69b6fe8947e5714ed4da101ddcf33e0f25",
		"0x8971997fb5d1a00599d30af8fe169a47e4d17d3efffefba52839537b89f54bb1",
		"0x5011902bd739d52f5d5c0e2bd382fc81350eb2c4f50aa0f1ce973e5fc41378e6",
		"0x3d371359b160afaeaba364bdcae5a6639699a513198ebb12fb3ee1f215e34e49",
		"0x496a267275fc974aac6d3f46921dd6a08441c9f1dbb9861dc5d8210a1d52053a",
		"0xf2e1e9548e5c15a9d21189ee2766770f0245b238499526fe538b505c7a159774",
		"0x48f59be041558af55f36dfc7ec5a7950d6482d10f8c0977b71a0b86e9c0a4767",
		"0xedeb92913dd239b28906688cae03d3790508c7df9b15fa6ef9abb4042c659985",
		"0xb4094d84262c1039240c21f4d8364b7e802940ba734573e9d8e7566573880c41",
		"0xdec2b48cfcd0273ad1f23404e0f9e9d3fc4fa1721c317d933484b9015858ba5d",
		"0x7dbd4f809ecdaa1ab4dcb40f6a24ddc4561fe264c3fe405e3dfcb7eb44f6275d",
		"0xe3cbb18271f2be064b2148bd15449f2cb92169cecee82ace180e38efef35ab99",
		"0xbfcf51ca0f25962a5a121234999a96818f94580fc7cfddabc0d0c6dbb98ba8a1",
		"0xfb2746579a4129d17c58a796832e30eacc3f329fec12babccc69c49601d93e06",
		"0x3ecffcca6aef54316e12532a4284829334a0d0ea98921ece4b8b296133a7d454",
		"0x56340e56c551897a8e4f6206b4dd9c581c242b39f3417a0460b6b1389c1db1cc",
		"0x2ad703afee0c00cd64f8a309df7caf56db476f50a3a63eb7b645f16afe347670",
		"0x499881041292f414309a062c4a0e50a32886a642ca5b18ce2f33892c01c59b6f",
		"0x169e05cb1c52f31bce923a62c3a54878e5009664ebdbda7d2c9c2eb8f9ef6c60",
		"0xeea4623936f85480e06e6fef37b4b8f6609df6060379c06826b53998c65db8a5",
		"0xd190fb5d73d6637e61219b9fec59d262e78ea2425e1b33ac90913d53f4198f57",
		"0x9887e54d2d5080cb822d36ba321f2f8e94ab86ec41202b32364330a66f3771f3",
		"0x27e029823dc6c29f811586f1c7494da3f98b21b66149b84aef9cecbf1b3b7d84",
		"0x1d13d33d4fb6fece8905aa1ca88b12b13d6657c0088e4aef925d9b841c8bd04e",
		"0x917b800fe9c34bb01f64466d72d16b17a1a8ee259fe75388728732fbd85efc56",
		"0xbdc439d615a794cd56f1e8db57443ecfb7a3aab3007b577c1384c326796e416e",
		"0x56f007b4a793aa597f2931918041d15e49c0f0df13d70e943abb1dd55443bdde",
		"0x8cbf6dd9f811bd995360e07cd473b03557b26edaa08c74a2b214b51a62a86add",
		"0x8d995b221fde9d68f694d67387fa5f7ba96cbf214859afdee35ee1259f40258a",
		"0xbfda7549339ec65a6194c590adae05afa8b151ac09e45d47089909baa2e85d0f",
		"0x327e90aacbb89dc399c16dda9bf9c678c34706a92a12f916937d51f06f7b77eb",
		"0xf209378bcbc64c73e8b7f79b458862fa5352ef98e3be6f35b939da11585f88bd",
		"0x48855ae3f55f7541100460044bdffff68d385c037e5d95467ecc7e9bf94717b4",
		"0xa8b35150d825158df3582e93b59697267da48e9c346fa0201b73740d494066c1",
		"0x39d071a2992e1f9cae76f063f6ac43a8c391f48e6dee0a747cb96fc484c8b1ad",
		"0x44f423e403ae230485e9252adad3a7919b15929264e74b153cc71d82e4aa4092",
		"0x0baa66376068c94bb0584b6ffd546b920fe3800bdf0738983b4664936bb77ab5",
		"0x16accf976a71e9ec5a529d4664adc78a3ddf54b8e5c9515c9b5cae0c510f84c5",
		"0x1edc90ec503856ab0ad0d0f94b23f32ed4fd4e0f40b135f742954e045d556cc9",
		"0x3349c201895a20ee27559947736cefaff2c8ea4e4f4596af993a32f96d574c7c",
		"0xe43760622c3706049c2e5f83286dfba2560f7f435f9a84c4a02580ab74ddcd3b",
		"0xcb5b181c33eb5799d13404b3aca7636f4b1754b1edfad6e032031ccef08c0a9b",
		"0xb68f242eb49ecb58129a46ce18f118a1da76f75753bf0b3c955dea35ca453b76",
		"0x88b6798175f3fdc2f6fb9a14c5d0223ce7f20f63bee3c37bcc3cdb19ef15314c",
		"0x8d1859fa479851868ce2c9e364402d393f80f7092aaa56b81669b110052d89c4",
		"0xaed286129bfe7b12eab109f95241eb00d869e951481077ff776169f2bdba5826",
		"0xd7a0a2519649a7ebd7e343291b245411e1d9280755ca844f80499b02ada3cc8e",
		"0xb72018ac17ae122c7af6d27e8bd10646980fa28dc31677de31ad4da676d93590",
		"0x6c8bdac67bcefc51948fa2312d0b96d63b50d64a1c8fc86c10524b5b0d0065bc",
		"0xe4dac2bcaed15966dade7969b3c846218333ea30aac8b40bb1f9aefaf450bf7f",
		"0xa58a4dabe4e61062381f3c5cfbdf9475da4d7753b95c2ceda3146a839317ef22",
	}
)

func CreateAddresses(N int) ([]string, []string) {
	keys := make([]string, 0, N)
	addrs := make([]string, 0, N)

	k := CreateAddressesRaw(N)
	for _, sk := range k {
		addr := crypto.PubkeyToAddress(sk.PublicKey)
		skHex := "0x" + common.Bytes2Hex(crypto.FromECDSA(sk))
		// Sanity check marshalling
		skTest, err := crypto.ToECDSA(crypto.FromECDSA(sk))
		if err != nil {
			panic(err)
		}
		_ = skTest
		keys = append(keys, skHex)
		addrs = append(addrs, addr.Hex())
	}
	return keys, addrs
}

func CreateAddressesRaw(N int) []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, 0, N)

	for i := 0; i < N; i++ {
		// WARNING= USES UNSECURE RANDOMNESS
		sk, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		keys = append(keys, sk)
	}
	return keys
}

func Airdrop(config *Config, value *big.Int) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(config.faucet.PublicKey)
	fmt.Printf("Airdrop faucet is at %x\n", sender)
	var tx *types.Transaction
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		fmt.Printf("error getting chain ID; could not airdrop: %v\n", err)
		return err
	}
	for _, addr := range config.keys {
		nonce, err := backend.PendingNonceAt(context.Background(), sender)
		if err != nil {
			fmt.Printf("error getting pending nonce; could not airdrop: %v\n", err)
			return err
		}
		to := crypto.PubkeyToAddress(addr.PublicKey)
		gp, _ := backend.SuggestGasPrice(context.Background())
		gas, err := backend.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     crypto.PubkeyToAddress(config.faucet.PublicKey),
			To:       &to,
			Gas:      30_000_000,
			GasPrice: gp,
			Value:    value,
		})
		if err != nil {
			fmt.Printf("error estimating gas: %v\n", err)
			fmt.Printf("estimating: from %v, to %v, gas %v, gasprice %v value %v", crypto.PubkeyToAddress(config.faucet.PublicKey), &to, 30_000_000, gp, value)
			return err
		}
		tx2 := types.NewTransaction(nonce, to, value, gas, gp, nil)
		signedTx, _ := types.SignTx(tx2, types.LatestSignerForChainID(chainid), config.faucet)
		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			fmt.Printf("error sending transaction; could not airdrop: %v\n", err)
			return err
		}
		tx = signedTx
		time.Sleep(10 * time.Millisecond)
	}
	// Wait for the last transaction to be mined
	if _, err := bind.WaitMined(context.Background(), backend, tx); err != nil {
		return err
	}
	return nil
}
