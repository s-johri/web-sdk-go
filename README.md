# PayU SDK for GO

The PayU SDK for GO enables you to easily work with APIs of PayU by easily integrating this SDK within your base system.
With our SDK, you do not need to worry about low level details for doing API integration and with few lines of code and a function calling, it can be done within few minutes.

## Features Supported
Following features are supported in the PayU GO web SDK:
 - Create Payment form.
 - Verify transactions, check the transaction status, transaction details, or discount rate for a transaction
 - Initiated refunds, cancel refund, check refund status.
 - Retrieve settlement details which the bank has to settle you.
 - Get information of the payment options, offers, recommendations, and downtime details.
 - Check the customer’s eligibility for EMI and get the EMI amount according to interest 
 - Pay by link genration

To get started with PayU, visit the [PayU developer documentation](https://docs.payu.in/).

API reference by family:
[transaction checks](https://docs.payu.in/reference/check-transaction-apis) ·
[refunds](https://docs.payu.in/reference/refund-apis) ·
[bin APIs](https://docs.payu.in/reference/bin-apis)

# Table of Contents
 1. [Usage](#usage)
 2. [Getting Started](#getting-started)
 3. [Documentation for various Methods](#documentation-for-various-methods)

## Usage

```shell
import (
  payu "github.com/payu-india/web-sdk-go"
)
```


## Getting Started

Please follow the [Usage](#usage) instruction and execute the following GO code for creating the instance of PayU Object:

```go

payu, err := payu.NewClient(
  <YOUR_MERCHANT_KEY>,
  <YOUR_MERCHANT_SALT>,
  <ENVIRONMENT>                 // Possible value = TEST/PROD
) 

```
## Documentation for various Methods 


Method |  Description
------------- | -------------
**GeneratePaymentForm**  | Genereate auto submit HTML form to intitiate transaction 
**VerifyPayment** | Provides the details of a transaction  
**GetTransactionDetails** | Provides the details of a transactions for a specfic timeperiod
**ValidateVPA** | Used to validate VPA of a user. 
**CancelRefundTransaction** | Initiate refunds. 
**CheckActionStatus** | Check the status of a refund.  
**GetNetbankingStatus** | Check downtime status of PGs. 
**GetIssuingBankStatus** | Check downtime through bin number. 
**CheckIsDomestic** | Check the bin information
**CreateInvoice** |  Used to create email and SMS invoice ( Pay by link ).
**ExpireInvoice** | Used to expire an existing invoice.
**EligibleBinsForEMI** |  Used for checking the card eligibilty for EMI through the bin number.
**GetEmiAmountAccordingToInterest** | Used to fetch interest accordign to Banks and tenure.
**GetSettlementDetails** |  Used to fetch settlement details for a particular date.
**GetCheckoutDetails** |  Used to fetch payment options, offers, eligibility, recommendations, and downtime details.
