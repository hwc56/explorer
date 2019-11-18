const PREFIX = ">";
const SUFFIX = "ago";

const TxType = {};
TxType.TRANSFER = 'Transfer';
TxType.BURN = 'Burn';
TxType.SETMEMOREGEXP = 'SetMemoRegexp';
TxType.CREATEVALIDATOR ='CreateValidator';
TxType.EDITVALIDATOR = 'EditValidator';
TxType.UNJAIL = 'Unjail';
TxType.DELEGATE = 'Delegate';
TxType.BEGINREDELEGATE = 'BeginRedelegate';
TxType.SETWITHDRAWADDRESS = 'SetWithdrawAddress';
TxType.BEGINUNBONDING = 'BeginUnbonding';
TxType.WITHDRAWDELEGATORREWARD = 'WithdrawDelegatorReward';
TxType.WITHDRAWDELEGATORREWARDSALL = 'WithdrawDelegatorRewardsAll';
TxType.WITHDRAWVALIDATORREWARDSALL = 'WithdrawValidatorRewardsAll';
TxType.SUBMITPROPOSAL = 'SubmitProposal';
TxType.DEPOSIT = 'Deposit';
TxType.VOTE = 'Vote';
TxType.ISSUETOKEN = 'IssueToken';
TxType.EDITTOKEN = 'EditToken';
TxType.MINTTOKEN = 'MintToken';
TxType.TRANSFERTOKENOWNER = 'TransferTokenOwner';
TxType.CREATEGATEWAY = 'CreateGateway';
TxType.EDITGATEWAY = 'EditGateway';
TxType.TRANSFERGATEWAYOWNER = 'TransferGatewayOwner';
TxType.REQUESTRAND = 'RequestRand';
TxType.ADDPROFILER = 'AddProfiler';
TxType.ADDTRUSTEE = 'AddTrustee';
TxType.DELETEPROFIKER = 'DeleteProfiler';
TxType.DELETETRUSTEE = 'DeleteTrustee';
TxType.CLAIMHTLC = 'ClaimHTLC';
TxType.CREATEHTLC = 'CreateHTLC';
TxType.REFUNDHTLC = 'RefundHTLC';
TxType.ADDLIQUIDITY = 'AddLiquidity';
TxType.REMOVELIQUIDITY = 'RemoveLiquidity';
TxType.SWAPORDER = 'SwapOrder';
TxType.TRANSFERS = 'Transfers';
TxType.WITHDRAWADDRESS = 'WithdrawAddress';
TxType.STAKES = 'Stakes';
TxType.GOVERNANCE = 'Governance';
TxType.DECLARATIONS = 'Declarations';



const ValidatorStatus = {};
ValidatorStatus.ACTIVE = 'active';
ValidatorStatus.JAILED = 'jailed';
ValidatorStatus.CANDIDATES = 'candidates';
ValidatorStatus.UNBONDED = 'Unbonded';
ValidatorStatus.UNBONDING = 'Unbonding';
ValidatorStatus.UNBONDED = 'Bonded';

const Denom = {};
Denom.IRISATTO = 'iris-atto';
Denom.IRIS = 'iris';
const ENVCONFIG = {};
ENVCONFIG.DEV = 'dev';
ENVCONFIG.QA = 'qa';
ENVCONFIG.STAGE = 'stage';
ENVCONFIG.TESTNET = 'testnet';
ENVCONFIG.MAINNET = 'mainnet';

const PARAMETER = {};
PARAMETER.EQUAL = 'eq';
PARAMETER.UNEQUAL = 'neq';

const CHAINNAME = 'iris';
const CHAINID = {};
CHAINID.IRISHUB = 'irishub';
CHAINID.FUXI = 'fuxi';
CHAINID.NYANCAT = 'nyancat';

const RADIXDENOM = {};

RADIXDENOM.IRISATTO = 'iris-atto';
RADIXDENOM.IRISATTONUMBER = '1000000000000000000';
RADIXDENOM.IRISMILLI = 'iris-milli';
RADIXDENOM.IRISMILLINUMBER = '1000000000000000';
RADIXDENOM.IRISMICRO = 'iris-micro';
RADIXDENOM.IRISMICRONUMBER = '1000000000000';
RADIXDENOM.IRISNANO = 'iris-nano';
RADIXDENOM.IRISNANONUMBER = '1000000000';
RADIXDENOM.IRISPICO = 'iris-pico';
RADIXDENOM.IRISPICONUMBER = '1000000';
RADIXDENOM.IRISFEMTO = 'iris-femto';
RADIXDENOM.IRISFEMTONUMBER = '1000';
RADIXDENOM.IRIS = 'iris';
RADIXDENOM.IRISNUMBER = '1';

const TRANSACTIONMESSAGENAME = {};
TRANSACTIONMESSAGENAME.TXTYPE = 'TxType:';
TRANSACTIONMESSAGENAME.FROM = 'From:';
TRANSACTIONMESSAGENAME.AMOUNT = 'Amount:';
TRANSACTIONMESSAGENAME.TO = 'To:';
TRANSACTIONMESSAGENAME.OWNER = 'Owner:';
TRANSACTIONMESSAGENAME.MEMOREGEXP = 'MemoRegexp:';
TRANSACTIONMESSAGENAME.OPERATORADDRESS = 'Operator Address:';
TRANSACTIONMESSAGENAME.MONIKER = 'Moniker:';
TRANSACTIONMESSAGENAME.IDENTITY = 'Identity:';
TRANSACTIONMESSAGENAME.SELFBONDED = 'Self-Bonded:';
TRANSACTIONMESSAGENAME.OWNERADDRESS = 'Owner Address:';
TRANSACTIONMESSAGENAME.CONSENSUSPUBKEY = 'Consensus Pubkey:';
TRANSACTIONMESSAGENAME.COMMISSIONRATE = 'Commission Rate:';
TRANSACTIONMESSAGENAME.WEBSITE = 'Website:';
TRANSACTIONMESSAGENAME.DETAILS = 'Details:';
TRANSACTIONMESSAGENAME.SHARES = 'Shares:';
TRANSACTIONMESSAGENAME.TOSHARES = 'Shares: ';//此处有空格
TRANSACTIONMESSAGENAME.ENDTIME = 'End Time:';
TRANSACTIONMESSAGENAME.NEWADDRESS = 'New Address:';
TRANSACTIONMESSAGENAME.ORIGINALADDRESS = 'Original Address:';
TRANSACTIONMESSAGENAME.PROPOSER = 'Proposer:';
TRANSACTIONMESSAGENAME.TITLE = 'Title:';
TRANSACTIONMESSAGENAME.INITIALDEPOSIT = 'Initial Deposit:';
TRANSACTIONMESSAGENAME.DESCRIPTION = 'Description:';
TRANSACTIONMESSAGENAME.PROPOSALTYPE = 'Proposal Type:';
TRANSACTIONMESSAGENAME.DEPOSITOR = 'Depositor:';
TRANSACTIONMESSAGENAME.PROPOSALID = 'Proposal ID:';
TRANSACTIONMESSAGENAME.DEPOSIT = 'Deposit:';
TRANSACTIONMESSAGENAME.VOTER = 'Voter:';
TRANSACTIONMESSAGENAME.OPTION = 'Option:';
TRANSACTIONMESSAGENAME.FAMILY = 'Family:';
TRANSACTIONMESSAGENAME.SOURCE = 'Source:';
TRANSACTIONMESSAGENAME.GATEWAY = 'Gateway:';
TRANSACTIONMESSAGENAME.SYMBOL = 'Symbol:';
TRANSACTIONMESSAGENAME.CANONICALSYMBOL = 'Canonical Symbol:';
TRANSACTIONMESSAGENAME.NAME = 'Name:';
TRANSACTIONMESSAGENAME.DECIMAL = 'Decimal:';
TRANSACTIONMESSAGENAME.SYMBOLMINALIAS = 'SymbolMinAlias:';
TRANSACTIONMESSAGENAME.INITIALSUPPLY = 'InitialSupply:';
TRANSACTIONMESSAGENAME.MAXSUPPLY = 'MaxSupply:';
TRANSACTIONMESSAGENAME.MINTABLE = 'Mintable:';
TRANSACTIONMESSAGENAME.TOKENID = 'TokenId:';
TRANSACTIONMESSAGENAME.ORIGINALOWNER = 'Original Owner:';
TRANSACTIONMESSAGENAME.NEWOWNER = 'New Owner:';
TRANSACTIONMESSAGENAME.BLOCKINTERVAL = 'Block Interval:';
TRANSACTIONMESSAGENAME.REQUESTID = 'Request ID:';
TRANSACTIONMESSAGENAME.RANDHEIGHT = 'Rand Height:';
TRANSACTIONMESSAGENAME.HASHLOCK = 'Hash Lock:';
TRANSACTIONMESSAGENAME.TIMELOCK = 'Time Lock:';
TRANSACTIONMESSAGENAME.TIMESTAMP = 'Timestamp:';
TRANSACTIONMESSAGENAME.EXPIRYHEIGHT = 'Expiry Height:';
TRANSACTIONMESSAGENAME.CROSSCHAINREVEIVER = 'Cross-chain Receiver:';
TRANSACTIONMESSAGENAME.SECRET = 'Secret:';
TRANSACTIONMESSAGENAME.PARAMETER = 'Parameter:';
TRANSACTIONMESSAGENAME.SOFTWARE = 'Software:';
TRANSACTIONMESSAGENAME.VERSION = 'Version:';
TRANSACTIONMESSAGENAME.SWITCHHEIGHT = 'Switch Height:';
TRANSACTIONMESSAGENAME.TRESHOLD = 'Treshold:';
TRANSACTIONMESSAGENAME.ADDRESS = 'Address:';
TRANSACTIONMESSAGENAME.ADDRESSBY = 'Added By:';

const SUBMITPROPOSALTYPE = {};
SUBMITPROPOSALTYPE.SUBMITSOFTWAREUPGRADEPROPOSAL = 'SubmitSoftwareUpgradeProposal';
SUBMITPROPOSALTYPE.SUBMITTXTAXUSAGEPROPOSAL = 'SubmitTaxUsageProposal';
SUBMITPROPOSALTYPE.SUBMITTOKENADDITIONPROPOSAL = 'SubmitTokenAdditionProposal';
SUBMITPROPOSALTYPE.SUBMITPROPOSAL = 'SubmitProposal';

export default {
  PREFIX,
  SUFFIX,
  TxType,
  ValidatorStatus,
  Denom,
  ENVCONFIG,
  CHAINNAME,
  PARAMETER,
  CHAINID,
  RADIXDENOM,
  TRANSACTIONMESSAGENAME,
  SUBMITPROPOSALTYPE
};
