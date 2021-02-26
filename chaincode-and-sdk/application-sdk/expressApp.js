const express = require('express')
const app = express()
const bodyParser = require('body-parser');
app.use(bodyParser.json());

const port = 3000

let contract;

const initialize = async () => {
  const { Gateway, Wallets } = require('fabric-network');
  const FabricCAServices = require('fabric-ca-client');
  const path = require('path');
  const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('../../test-application/javascript/CAUtil.js');
  const { buildCCPOrg1, buildWallet } = require('../../test-application/javascript/AppUtil.js');

  const channelName = 'mychannel';
  const chaincodeName = 'basic';
  const mspOrg1 = 'Org1MSP';
  const walletPath = path.join(__dirname, 'wallet');
  const org1UserId = 'appUser';
  
  const ccp = buildCCPOrg1();
  const caClient = buildCAClient(FabricCAServices, ccp, 'ca.org1.example.com');
  const wallet = await buildWallet(Wallets, walletPath);
  await enrollAdmin(caClient, wallet, mspOrg1);
  await registerAndEnrollUser(caClient, wallet, mspOrg1, org1UserId, 'org1.department1');

  const gateway = new Gateway();

  await gateway.connect(ccp, {
	wallet,
	identity: org1UserId,
	discovery: { enabled: true, asLocalhost: true }
  });

 gateway.getNetwork(channelName).then(network => {
   contract = network.getContract(chaincodeName);
   contract.submitTransaction('InitLedger').then(() => console.log('Init ledger')).catch(error => console.log(`Successfully caught the error: \n    ${error}`));
 });
}


/** GetAllCarsByColor **/
app.get('/get-all-cars-by-color', (req, res) => {
  contract.evaluateTransaction('GetAllCarsByColor', req.query.color).then(result => res.send(`Result: ${result.toString()}`)).catch(error => res.send(`Successfully caught the error: \n    ${error}`));
})

/** GetAllCarsByColorAndOwner **/
app.get('/get-all-cars-by-color-and-owner', (req, res) => {
  contract.evaluateTransaction('GetAllCarsByColorAndOwner', req.query.color, req.query.owner).then(result => res.send(`Result: ${result.toString()}`)).catch(error => res.send(`Successfully caught the error: \n    ${error}`));
})

/** CreateFailure **/
app.post('/create-failure', (req, res) => {
  contract.submitTransaction('CreateFailure', req.body.id, req.body.failureID, req.body.desc, req.body.price).then(() => res.send('Request sent')).catch(error => res.send(`Successfully caught the error: \n    ${error}`));
})

/** TransferOwnership **/
app.post('/transfer-ownership', (req, res) => {
   contract.submitTransaction('TransferOwnership', req.body.id, req.body.acceptFailures, req.body.newOwner).then(() => res.send('Request sent')).catch(error => res.send(`Successfully caught the error: \n    ${error}`));
})

/** RepairFailure **/
app.post('/repair-failure', (req, res) => {
   contract.submitTransaction('RepairFailure', req.body.id, req.body.failure).then(() => res.send('Request sent')).catch(error => res.send(`Successfully caught the error: \n    ${error}`));
})

/** ChangeColor **/
app.post('/change-color', (req, res) => {
   contract.submitTransaction('ChangeColor', req.body.id, req.body.color).then(() => res.send('Request sent')).catch(error => res.send(`Successfully caught the error: \n    ${error}`));
})

/** RUN **/
app.listen(port, () => {
 initialize()
 console.log(`Example app listening at http://localhost:${port}`)
})
