package gossip_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gohornet/hornet/pkg/metrics"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/p2p"
	"github.com/gohornet/hornet/pkg/protocol/gossip"
	"github.com/gohornet/hornet/pkg/testsuite"
	iotago "github.com/iotaledger/iota.go/v2"
)

const (
	MinPoWScore   = 100.0
	BelowMaxDepth = 15
)

func TestMsgProcessorEmit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownSignal := make(chan struct{})
	defer close(shutdownSignal)

	te := testsuite.SetupTestEnvironment(t, &iotago.Ed25519Address{}, 0, BelowMaxDepth, MinPoWScore, false)
	defer te.CleanupTestEnvironment(true)

	// we use Ed25519 because otherwise it takes longer as the default is RSA
	sk, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	n, err := libp2p.New(
		ctx,
		libp2p.Identity(sk),
		libp2p.ConnectionManager(connmgr.NewConnManager(1, 100, 0)),
	)
	require.NoError(t, err)

	serverMetrics := &metrics.ServerMetrics{}

	manager := p2p.NewManager(n)
	go manager.Start(shutdownSignal)

	service := gossip.NewService(protocolID, n, manager, serverMetrics)
	go service.Start(shutdownSignal)

	networkID := iotago.NetworkIDFromString("testnet4")

	processor, err := gossip.NewMessageProcessor(te.Storage(), gossip.NewRequestQueue(), manager, serverMetrics, &gossip.Options{
		MinPoWScore:       MinPoWScore,
		NetworkID:         networkID,
		BelowMaxDepth:     BelowMaxDepth,
		WorkUnitCacheOpts: testsuite.TestProfileCaches.IncomingMessagesFilter,
	})
	require.NoError(t, err)

	msgData := `{
		  "networkId": "9466822412763346725",
		  "parentMessageIds": [
			"42e53f6bc0ecaf69f0f32dfbd838a0f96396c09b92e53225784ee9d269671939",
			"77cfab5b59a894bd5992b303e93c126191257c025136c610dd351b06863a9ee3",
			"912f97dd2b76ee450bf8495f2c1e47d4255482a5d62aba3bb30e5e2555dea164",
			"a0fe3e8192f0bc4270dd38b3ff08c673cb2253d75891a54bbe534540ae770623"
		  ],
		  "payload": {
			"type": 2,
			"index": "362527d26c54bfc00049f46049593e7d5ff4bc0856e4a235347614b72503889b",
			"data": "064d59394d414d41514b394541544f595a414b594a43395a5551564b4744415645534c544a4c5a4e594247465050514d45524b54584d524e4b584241465851575947554c445652555655414f424a48464b54455546454958524a58395956545a4f58394a394f4153524c505a4347485958544a5654474d52504d44434243414b42395a435642594b41594251443949424a444f46474d4d4d595a58444359574b4c4147594f39545a565755504857465945444e4d4b39565044414c4354535a45594c434c464e4e565a505647434b485a54524555525249494a59444239445849575a4357474f4e59524c434c4145444639504f553950554c544c514d5955425741454b58464e4553494f45544443425539474942464e4147414b4e544e4251454f424d4856484f52545a5a4949414f564d514c43465a5349524f5143584a4742564e554f4d524939535458455a4141555a5149494e455658554657394d4956514c4755583945574159584e4e464e594d514f41414939544a53524e554b5149414442484947443958504e4441574257504a514c4c4442464441504457504e47444a4357464c43535058444e495a574b4e425556564e4d4e544a4b58424657505944554f485342584b4d565a4f44394d51584e5051515348555a48524751515249504154504846584b4d4a4359524a49414859474c4747495a43455955574e474b594a4d4b39515250444c594c4152515456424d43464d544948514744474856504a445857395544434945584d524647434d51425743594a484f5454424f534b51434f434e5746535359464244464254595551444f594b53574d4b424e4452465a52574841475a4a5a4250504758495559565a5850414b4e5841594d5142554d4f49553946554d4d46474d45485557504f495a4a47534e4d434f59494a49555041435542555a494448515445394a4b584848544a4d505150484b544f5a5954495655415742424e4a5444424a50534846455945474c4d56555241574f4359484c5a394339554b4b504949484b445457584753415158455839475753425951423951504443444a525439575842454750484d49505354423951565a545156544b4c534d564a51395343395250444c505252514c584945525a504e52484e46534e524b525043424950504b53495a49444f444e565757594a504c393958495541494e4c574d5745445344425757535248475641495a484153464b434c575751585149544b4c4c41534d41524e515050544957484b54505a584e394a4c5758564a59454645544243515557554e4a465641594d554d4654435859564d564c554f45414a4c57574b3956485a4645394d57474c48515444515847394b5942544e58534d545845584b5652425342394639514e4655514e444a5156464845574f51544852554a43434e59594a47455744534e55424b5a443946495552574a48474f4a5a5354504b4c4758584359425656425439594651444a524957564f444c5a464644555745474657494f4b584d42484150535559425457564c50543953565a43474d39554846584e425553505a4148424850555a5a5355504a494d5948444f4e5856505258414e485255445a4d41455044444c4a4d5a534f594e4b554e45414b47414d5a47474c41475558574243435047474d434c3955585a43585755484c585750434c5746485844525859594a5756424d4b4a5758544f4b454e514953564b584b56514855595044534e4f4b4c45394751395350485054464450465143495851594c534a51454b585a5941505044434d51565550434347394b59544e4c584d4f4a5359394d4d444957455a42414f474c4a4553394b4754504b4349534358474b5354474a4d455450464b434e424f4342515839484d4745534f48585250394643454e414d444a42484b475257514f49555952484d48544c394957484c484d484b4149554a455148394c5350564a54594950445a394e565239425a475839464941555046394453474f5255415757394c534249454e555041454c4b4b48424e4f57524a424d45514658414e4e4442575156585a545a4b5151495a435944425351595246454e39424f5552415257575239414847574147415250525649585948474d39424a4e59574d5148545a4b464c50524b565a4848524f434b5548513959425345414c56544848434452444154485349464858454f58464845544a41464e4f505555485750564f564f4f5841454f394251514647545a39433955554356584639433943575041484242584a4a564644584f4857394447524a435653395957564e544a584c4c59504a4d394456484a415545434c4c59414a444c394153545a444151485756484c4d585739544c564c4347394c4a5539584d55585745584e55424950524239454356515a4747535557514e4759584f4742424441505056574f5839584c48554556413955534f444839434a485349544657444a4255565355414c4a564752565442534b47595a4d4156434d4b445554414e4d5944394d4a42524559554c4c52574255534a544347414b51435641485a584554584d4748594d4b423958484b5456565644465a545a515539534d4852544e425a5242414459594f57425050434a474852484e50434c3941414d584d504c4b58454c4e445343485457444357444c565a445757394f444854595945534e495a4f414d5758494456445a58465555415551454453515a54464e464c5a485655454f584648545a43515942544148555a5754574747585a5446564b48595a50545645474e47434d44575849595652454f4954515a4c44414f465844554d4b5154444442424b46545151464f42465a57514b544b4f564e584f5743425555504445544d43393946454d4542535a45554a4745575448514e474f474b4a514d534456585244434953545a4f5539494f445439435a5255444e504249434e4e5847475952584559504d565859575452564d4a425a455043484c424358434a52394541455a435a414c4e47474d5947514a394e534441484e54394f55434c474d4c474f4c52495a4d4654434939494e55524f574a4b4241465749434d545a44395742425839414957494950484a5557585846484f414b4b4f4c504d504e4e4a4f504e4f4f52475749555956555a4241474445434d4f524d51464d53584e534d43414a5352514249534e46434c39494952485151584f42544b574f4a5a56534c455449445a4f4d445a4d5a4b4646504c4d4f48514b4b4745524d4d58454f4c424e434b39544554585539554a4c424a4b574439593956475251555a5452464d554242584e56514c4c5757425749394a48585a4d4e5951434c5539475143474555544f41444e535054544842425a4b5550564f5446594944475a4c584d5545394a554a534a4b565a455157555a50564f4a4f54424744394c504c5950564b484454554c39434339515341484f56414f524e5350394d4a5555594d4a4a52454a574f494c47534d47454f39594c54564657495941494e464a58474544574b57554d424458534f434d4e56535152454848504c414e4e575845414f51564a5156494a58424847484847564c4c4f47565a5a474b5046434758444b48554f4b4d444e44534b55444c5649595650514b59594e48574a53514150574a495a454347475244565856394f395246565a434844564d415555474b4348524c5542595749414d435843494b504348514d44465449575143464f4d444b414c585659504b5045514347455354514c52545a524153574a454f41594c4a534c56575055554752505351535245465251534d4b58584744444b43574b484a4547493955505a474757454b48494c4e4f4b3939494757454554394342575747545a4f51434254594a524c4752434e424350485a475647594b47534f5742434c4d4b4c4b455342574a50474f514d4e57573949395a4e5a4d4b4a564954584c4b4c39544c47514a57434c544c435051514b4f45395739414b4a4e555849394845545155574c4a554352504f4655444956544d5352524743524751554b5657504f5a485443525a4b53475a415a5244574c58564c57524d44534e574c585a48574d4f4c49434a42415a4f5548464a42465649525254554d4a43494e4544524b54424f5a514b574d504b514b46544751525152524d4754524949464b4858454b5a4b5945494555394f514d4757414d5a4344454a4c59534439494657414e4c5346594a564f59514142423954475248564a4f504e5a5855504f494b4e5652414e48564942544655514f5655423945474d4d4259554f554a394751424b4f4b4c505752594d394d474b4a47434e5146554a56494c59454b4745414d5145395a594f4b52544142394c5547415a51435a4e4f3942444742464e534f484352394a4f4d4f525941425841445a4756444a4d484e4945494239444a555259424e5a444159574f504c5453435544594257454c594b4c425a5754575247525952424a51495a4d5039595454484e56394442504c5954494a4e4e485a465759574d3954544245534d444e554f594e4c425645585852454254584245564c484f585353584f415356474a4b4f5a4e474a4e4c59444a475450514d4e5a584e504b50394c485a5451445549394e59494f4f434e433939474551424b46454c4f574f464a4b4d5845494e56454d4552393952534c4f5950595656424f444f524a424a50524d544c4d504d58434a595353574f47595a5854574c5a4b534741495543434239473947564258575659494c4f4358565158444e54414741435856544f4c4d455a4656494b415a44595743473949434f39515a4b57414857495541564b4448484646395657504854534141475739475a425752424c44445446585756514a4d39494f594c444c534e474356565745595657544f4552504457524d464e3959504245544e5244514f53505339534948544d414f504c49574a414f594c544d48495145523955564f575747414e4245504d483941564e4c444a42484f5156465545425239444752534b4242544146514d47424a575942564b4d4d4f4b4f4d454b484c5a4b534d56524c524b395842424d544a3958454c4d5a484146574d5045544d434a5155485a49524b4b4c4144444f4c4b4b455141544e4c4c4e484d4c5045394449545553525247454c535945474e424f454f4f464f43485255435257514c4b3947454f5646534e56474d47394541414d47593941415a5a564a4b4a484c49435257474245445153584656475054595648494e59464a4b564a5156464a45584a4a43554f4a47394e4559594558524a5a434a5a4c5a415152444541424156425450585658464b465053544659475747484e5151514539445449534d4b544d504f395948425548495a58484e5945534c39464a5356434842475a5357584b47594b444847434a52444452414c4d56585449584a39495553484d484854414c41414e55594942504556505752554e4e4b4d415a59564e4b47434d42465759394258554f4541554742395050474e4e5756495057524152464846584a4f494e494a44414a45414d5550444e594b5945444856504f474e594952444b47414545445648425a4947424b484c4741524d4839494d5443584e4346514a45434f53515846464c58414c47465350564e534e4a5447524c445355495055475148514247524d584b394e5148554458454648564353414b5243484c55424949454f4153555451464c43514945535a56524448505758474b564f54584d395852554e4d4d555953515244464f49584a53564457425150414d54554e4c4c4d4a545844584b594d585545445a5a4e464a58504c42475939394c41524739594639424150484d554c45514c4a39424c50504d4c5a5053585750584a594b5046524d484743534f47494a52444e584e5145415a434f4e555051585a4d5743424b464341534251434a55555a5a504b4f544243474b585742494b4f4e47554a46535a42565a514651515342394f4d524f394151415056594c49424554454a504d4957594a5454525759414c504146555a445354585346464f495042424758484d43455a51585842444c544a424a57535847544d474b544b574b4b50594d49473959574f4a44474d4e5455484d574352544e454e48524a584c4c504254465a58565a4c524e584f4a5a4849525539584c4d4a504e574e4c574d4a564b454b564b51393956425a524659534e4b514654444f4953434152544d5052414954415342424f43465a44484953455041395153484c54584656595a47555a5058395250434139475a484a495a"
		  },
		  "nonce": "13835058055282328477"
		}
		`

	msg := &iotago.Message{}
	assert.NoError(t, json.Unmarshal([]byte(msgData), msg))

	message, err := storage.NewMessage(msg, iotago.DeSeriModePerformValidation)
	assert.NoError(t, err)

	// should fail because parents not solid
	err = processor.Emit(message)
	assert.Error(t, err)

	// set valid parents
	msg.Parents = iotago.MessageIDs{[32]byte{}}

	// pow again, so we have a valid message
	err = te.PoWHandler.DoPoW(msg, nil, 1)
	assert.NoError(t, err)

	// need to create a new message, so the iotago message is serialized again
	message, err = storage.NewMessage(msg, iotago.DeSeriModePerformValidation)
	assert.NoError(t, err)

	// should not fail
	err = processor.Emit(message)
	assert.NoError(t, err)

	// set wrong network ID
	msg.NetworkID = 1

	// pow again, so we have a valid message
	err = te.PoWHandler.DoPoW(msg, nil, 1)
	assert.NoError(t, err)

	// need to create a new message, so the iotago message is serialized again
	message, err = storage.NewMessage(msg, iotago.DeSeriModePerformValidation)
	assert.NoError(t, err)

	// message should fail because of wrong network ID
	err = processor.Emit(message)
	assert.Error(t, err)

	// set valid network ID again
	msg.NetworkID = networkID

	// pow again, so we have a valid message
	err = te.PoWHandler.DoPoW(msg, nil, 1)
	assert.NoError(t, err)

	// need to create a new message, so the iotago message is serialized again
	message, err = storage.NewMessage(msg, iotago.DeSeriModePerformValidation)
	assert.NoError(t, err)

	// should not fail
	err = processor.Emit(message)
	assert.NoError(t, err)

	// set wrong nonce
	msg.Nonce = 123

	// need to create a new message, so the iotago message is serialized again
	message, err = storage.NewMessage(msg, iotago.DeSeriModePerformValidation)
	assert.NoError(t, err)

	// should fail because of wrong score
	err = processor.Emit(message)
	assert.Error(t, err)
}
