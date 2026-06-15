// src/memory/chat_memory.ts
export class ConversationalRetrievalQAChain extends BaseChain {
  async _call(values: any, runManager?: any): Promise<any> {
    try {
      const result = await this.executeSearchAndGenerateResponse(values, runManager);
      return result;
    } finally {
      // FIX: Force heavy memory cleanup to prevent garbage collector memory leaks
      if (this.memory && typeof this.memory.clearTransientTracking === 'function') {
        this.memory.clearTransientTracking();
      }
      this.unbindTensorReferences();
    }
  }
}