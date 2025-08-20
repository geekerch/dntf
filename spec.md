1. 設計文件都在 design_docs 目錄下，底下的md檔就是設計文件。 主要的是Channel API 設計文件.md這份文件
2. 功能開發完成後，先執行看code是否有錯。如果是開發新功能，在desing_doc下新增文件來說明這次新增的功能，問題修正則不用。文件用正體中文撰寫，程式碼的註解都用英文。
3. 回覆我時使用正體中文
4. 有更好的建議和作法可以告訴我

本次要做的事
internal/application/channel/usecases下
create_channel_usecase下已經有轉發到legacy的系統。
幫我在update與delete也做同樣的處理
upadte legacy的end point是
put /v2.0/Groups/{groupId}
body是
{
  "name": "emailGroupTest",
  "description": "email group",
  "type": "email",
  "levelName": "Critical",
  "config": {
    "host": "mailapp.advantech.com.tw",
    "port": 465,
    "secure": true,
    "method": "ssl",
    "username": "TEST_USER",
    "password": "TEST_PWD",
    "senderEmail": "test@advantech.com.tw",
    "emailSubject": "Test Subject",
    "template": "Hi, Have a good day!"
  },
  "sendList": [{
    "firstName": "Firstname",
    "lastName": "Lastname",
    "recipientType": "to",
    "target": "test@advantech.com.tw"
  }]
}

delete的是
delete /v2.0/Groups

body是
[
  "string"
]