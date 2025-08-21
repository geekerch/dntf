1. 設計文件都在 design_docs 目錄下，底下的md檔就是設計文件。 主要的是Channel API 設計文件.md這份文件
2. 功能開發完成後，先執行看code是否有錯。如果是開發新功能，在desing_doc下新增文件來說明這次新增的功能，問題修正則不用。文件用正體中文撰寫，程式碼的註解都用英文。
3. 回覆我時使用正體中文
4. 有更好的建議和作法可以告訴我

本次要做的事
幫我寫測試用例與測試
測試internal/presentation/nats/handlers
cqrs的不用測試，要測試的是channel，message與template的功能

. template的測試，要包含crud。並且在修改後是否會同步給channel
. channel的測試除了crud外，channel建立出來的id是會轉到舊系統上建立，舊系統叫groupId。 要測試crud時，舊系統是否也有同步更動。

. 舊系統可以用這隻api來查詢所有的groupGet /v2.0/Groups?count=1000&index=1&desc=false
. message要測試是否能寄。template有沒有帶，variable有沒帶這些都要測，還有改件人，cc，to，bcc都要有測。
 smtp的設定可以參考
 "config": {
      "host": "smtp.gmail.com",
      "port": 465,
      "secure": true,
      "method": "ssl",
      "username": "chienhsiang.chen@gmail.com",
      "password": "tlrqyoxptgjbbatn",
      "senderEmail": "chienhsiang.chen@gmail.com"
    },
	
有不完的地方可以補充。重點要完整的測試