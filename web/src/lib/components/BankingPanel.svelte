<script>
  import { createEventDispatcher, tick } from "svelte";
  import { token } from "../authStore";

  export let accounts = [];
  const dispatch = createEventDispatcher();

  let chatContainer;
  let userInput = "";
  let loading = false;

  let messages = [
    { role: "assistant", text: "Hola. Soy tu asistente de Nexora. ¿En qué puedo ayudarte?", type: "text" }
  ];

  async function scrollToBottom() {
    await tick();
    if (chatContainer) chatContainer.scrollTop = chatContainer.scrollHeight;
  }

  function addMessage(role, text, type = "text", extra = null) {
    messages = [...messages, { role, text, type, extra }];
    scrollToBottom();
  }

  async function handleSend() {
    if (!userInput.trim() || loading) return;
    
    const text = userInput.trim();
    userInput = "";
    
    addMessage("user", text);
    loading = true;

    try {
      const res = await fetch("http://localhost:8080/api/chat", {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${$token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          message: text,
          
          history: messages.slice(-6).map(m => ({
            role: m.role,
            content: m.text
          }))
        })
      });

      const data = await res.json();
      
      if (res.ok) {
        let reply = data.reply || "";
        
       
        reply = reply.replace(/\(tool_call\)[\s\S]*?(\]|\s|$)/g, "");
        reply = reply.replace(/\(function=.*?\)/g, "");

        addMessage("assistant", reply);
        
        
        const lower = reply.toLowerCase();
        if (lower.includes("✅") || lower.includes("exitoso") || lower.includes("realizado") || lower.includes("balance")) {
          
          setTimeout(() => {
            dispatch("refresh");
          }, 400);
        }
      } else {
        addMessage("assistant", "❌ Hubo un problema con la petición.");
      }
    } catch (err) {
      addMessage("assistant", "❌ Error de conexión con el Agente Nexora.");
    } finally {
      loading = false;
    }
  }

 
  function selectAccount(acc) {
    userInput = `Usa mi cuenta ${acc.account_number}`;
    handleSend();
  }
</script>

<div class="chat-panel">
  <header class="hd">
    <div class="gem">◈</div>
    <div class="hd-title">Nexora AI Assistant</div>
    <div class="status-dot"></div>
  </header>

  <div class="messages" bind:this={chatContainer}>
    {#each messages as msg}
      <div class="msg-wrapper {msg.role}">
        <div class="msg-bubble">
          {#if msg.type === "text"}
            {msg.text}
          {:else if msg.type === "account_selector"}
            <p>{msg.text}</p>
            <div class="btn-grid">
              {#each accounts as acc}
                <button on:click={() => selectAccount(acc)}>{acc.account_number}</button>
              {/each}
            </div>
          {:else if msg.type === "balance_list"}
            <div class="data-box">
              {#each (msg.extra || []) as acc}
                <div class="data-row"><span>{acc.account_number}</span> <b>${acc.balance.toFixed(2)}</b></div>
              {/each}
            </div>
          {:else if msg.type === "history_table"}
            <div class="data-box scrollable">
              <table>
                {#each (msg.extra || []).slice(0, 6) as tx}
                  <!-- svelte-ignore node_invalid_placement_ssr -->
                  <tr>
                    <td class={tx.type === 'CREDIT' ? 'up' : 'down'}>{tx.type === 'CREDIT' ? '↓' : '↑'}</td>
                    <td>${(tx.amount/100).toFixed(2)}</td>
                    <td class="date">{new Date(tx.date).toLocaleDateString()}</td>
                  </tr>
                {/each}
              </table>
            </div>
          {/if}
        </div>
      </div>
    {/each}
    {#if loading}<div class="typing">...</div>{/if}
  </div>

  <form class="input-area" on:submit|preventDefault={handleSend}>
    <input bind:value={userInput} placeholder="Ej: 'depositar' o 'ver saldo'..." />
    <button type="submit">➤</button>
  </form>
</div>

<style>
  /* Se mantienen tus estilos profesionales anteriores */
  .chat-panel { background: #fff; border-radius: 24px; border: 1px solid #e2e8f0; height: 500px; display: flex; flex-direction: column; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
  .hd { background: #0a1628; padding: 12px 18px; display: flex; align-items: center; gap: 10px; color: white; }
  .hd-title { font-weight: 700; font-size: 0.8rem; flex: 1; letter-spacing: 0.5px; }
  .gem { color: #c9a84c; font-size: 1.1rem; }
  .status-dot { width: 7px; height: 7px; background: #22c55e; border-radius: 50%; box-shadow: 0 0 5px #22c55e; }
  .messages { flex: 1; overflow-y: auto; padding: 18px; background: #f8fafc; display: flex; flex-direction: column; gap: 14px; }
  .msg-wrapper.user { justify-content: flex-end; }
  .msg-bubble { max-width: 85%; padding: 10px 14px; border-radius: 16px; font-size: 0.78rem; line-height: 1.5; }
  .assistant .msg-bubble { background: white; border: 1px solid #e2e8f0; color: #1e293b; border-bottom-left-radius: 4px; }
  .user .msg-bubble { background: #0a1628; color: white; border-bottom-right-radius: 4px; }
  .btn-grid { display: grid; gap: 6px; margin-top: 10px; }
  .btn-grid button { background: #f1f5f9; border: 1px solid #cbd5e1; padding: 7px; border-radius: 10px; font-size: 0.7rem; font-weight: 700; cursor: pointer; transition: 0.2s; color: #475569; }
  .btn-grid button:hover { background: #0a1628; color: white; border-color: #0a1628; }
  .data-box { margin-top: 8px; border-top: 1px solid #f1f5f9; padding-top: 8px; }
  .data-box.scrollable { max-height: 150px; overflow-y: auto; }
  .data-row { display: flex; justify-content: space-between; padding: 4px 0; font-size: 0.72rem; border-bottom: 1px solid #f8fafc; }
  table { width: 100%; border-collapse: collapse; font-family: monospace; }
  td { padding: 5px 2px; font-size: 0.7rem; }
  .date { color: #94a3b8; text-align: right; font-size: 0.65rem; }
  .up { color: #16a34a; font-weight: bold; }
  .down { color: #ef4444; font-weight: bold; }
  .input-area { padding: 12px; background: white; border-top: 1px solid #e2e8f0; display: flex; gap: 8px; }
  input { flex: 1; border: 1px solid #e2e8f0; padding: 9px 14px; border-radius: 12px; outline: none; font-size: 0.75rem; background: #f9fafb; transition: 0.2s; }
  button[type="submit"] { background: #0a1628; color: white; border: none; width: 36px; height: 36px; border-radius: 10px; cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 0.8rem; }
  .typing { font-size: 0.7rem; color: #94a3b8; padding-left: 10px; font-style: italic; }
</style>