1. 設計文件都在 design_docs 目錄下，底下的md檔就是設計文件。 主要的是Channel API 設計文件.md這份文件
2. 功能開發完成後，先執行看code是否有錯。如果是開發新功能，在desing_doc下新增文件來說明這次新增的功能，問題修正則不用。文件用正體中文撰寫，程式碼的註解都用英文。
3. 回覆我時使用正體中文

上次做完的事，但還沒commit。除錯與驗證中
domain中的channel．channelType是可以是像email，sms這些。我要加入一個功能，這個channelType是可以擴充實作的。例如email，也可以再實作slack...等。我要一個機制，定義好介面，讓其他人可以實作，並且可以掛在上來擴充而不用動到我的主程式

本次要做的事
發現到的問題，建立template時，他的channelType無法轉換，檢查與修正
{
  "ReqSeqId": "aaa",
  "data": {
    "name": "我的模板名稱2111",
    "channelType": "email",
    "subject": "這是一個主題",
    "content": "這是模板的內容，可以使用 {variable_name} 來表示變數。",
    "variables": [
      "variable_name",
      "另一個變數"
    ],
    "tags": [
      "行銷",
      "促銷"
    ],
    "settings": {
      "someSetting": "someValue",
      "anotherSetting": 123
    }
  }
}