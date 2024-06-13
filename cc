create account 

add name to it add filter fiat



 
{
  "name": "Demo wallet",
  "currency": "string",
  "beneficiaryId": "BNF-ZTEST-0DFUSJPF7A",
  "submitterEmail": "api-maker@zodia.io",
  "walletId": "ZTEST-NOBENF-849P8XKKAO"
}





==================

https://help.circle.com/s/article-page?articleId=ka0Un00000011rmIAA (circle registration)

https://help.circle.com/s/article-page?articleId=ka0Un00000011plIAA  (USDC deposit FAQs)

=================== DESIGN PORTAL ==================


Reports & statements
 -- Document type, Wallet, (From:xxx To:xxx)

Settings Page
 -- full Name
 -- Phone number
 -- change password
 -- my sessions (list all, delete all except, delete specific)
 -- theme (system, light, dark)
 -- permissions 




{
  "wallets": [
    {
      "walletId": "ZTEST-NOBENF-849P8XKKAO",
      "walletName": "Demo wallet",
      "walletCompany": "ZTEST",
      "createdBy": "api-maker@zodia.io",
      "status": "ACTIVE",
      "valid": true,
      "registeredComplianceKey": true,
      "blockchain": "string",
      "assetBalances": [
        {
          "currency": "string",
          "unitCurrency": "string",
          "availableUnitBalance": "641400",
          "ledgerUnitBalance": "641400",
          "pendingOutUnitBalance": "0",
          "pendingInUnitBalance": "0",
          "availableBalance": "0.006414",
          "ledgerBalance": "0.006414",
          "pendingOutBalance": "0",
          "pendingInBalance": "0"
        }
      ],
      "createdAt": "2021-08-03T18:17:57Z",
      "updatedAt": "2021-08-03T18:17:57Z"
    }
  ]
}


// transaction statistics
{
  "from": "2021-08-03T18:17:57Z",
  "to": "2021-08-03T18:17:57Z",
  "transferOutStats": {
    "CANCELLED BY MAKER": 110,
    "REJECTED BY AUTHORISER": 40,
    "FAILED": 367,
    "CONFIRMED": 20,
    "REJECTED BY SYSTEM": 126
  },
  "transferInStats": {
    "UNLOCKED": 994,
    "REJECTED BY SYSTEM": 37,
    "PENDING UNLOCK": 38
  }
}





