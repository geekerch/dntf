1. 設計文件都在 design_docs 目錄下，底下的md檔就是設計文件。 主要的是Channel API 設計文件.md這份文件
2. 功能開發完成後，先執行看code是否有錯。如果是開發新功能，在desing_doc下新增文件來說明這次新增的功能，問題修正則不用。文件用正體中文撰寫，程式碼的註解都用英文。
3. 都沒有問題後，把程式commit到git上，要寫上註解做了哪些處理
4. 回覆我時使用正體中文

本次要做的事
func (h *MessageNATSHandler) handleSendMessage(msg *nats.Msg) {
1. 改寫internal/infrastructure/presentation/message_nat_handler.go中的handleSendMessage。 在82行原本是調用response, err := h.sendUseCase.Execute(ctx, &request)。現在加上Forward這隻，改寫成先轉導到舊系統處理。目前Forward實做到一半，把他完成。
然後把我原本的channelId改成channelIds，實現可以帶多個channelId


