1. 設計文件都在 design_docs 目錄下，底下的md檔就是設計文件。 主要的是Channel API 設計文件.md這份文件
2. 功能開發完成後，先執行看code是否有錯。如果是開發新功能，在desing_doc下新增文件來說明這次新增的功能，問題修正則不用。文件用正體中文撰寫，程式碼的註解都用英文。
3. 回覆我時使用正體中文

本次要做的事
在/home/cch/eventcenter/nnnf/internal/application/message/usecases/send_message_usecase.go中
306行的回覆，應該是跟288行一樣是一個array。每一個要對應回應。
這個的總回覆是成功，裡面再包每個channel的處理狀況