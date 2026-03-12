<script>
  import { onMount, tick } from "svelte";
  import { fade, fly } from "svelte/transition";
  import { token, user, logout } from "../authStore";
  import { router } from "tinro";

  import AccountCard from "../components/AccountCard.svelte";
  import TransactionModal from "../components/TransactionModal.svelte";
  import HistoryModal from "../components/HistoryModal.svelte";
  import BankingPanel from "../components/BankingPanel.svelte";

  let accounts = [];
  let message = "";
  let chatHistory = [
    {
      role: "assistant",
      content: `Welcome to Nexora. I am your financial assistant powered by TigerBeetle. How can I help you today?`,
    },
  ];

  // LOADING STATES (UI/UX)
  let loadingChat = false;
  let isProcessingModal = false;

  // NOTIFICATION SYSTEM (Toasts)
  let notification = { show: false, message: "", type: "success" };
  let notificationTimeout;

  function showNotification(msg, type = "success") {
    notification = { show: true, message: msg, type };
    clearTimeout(notificationTimeout);
    notificationTimeout = setTimeout(() => {
      notification.show = false;
    }, 4000);
  }

  // Modal Controllers
  let showModal = false;
  let modalType = "";

  let showHistory = false;
  let historyTransactions = [];
  let historyAccountLabel = "";
  let loadingHistory = false;

  $: displayName =
    (typeof $user === "string" ? $user : $user?.fullname || $user?.name) ||
    "User";

  function handleLogout() {
    logout();
    router.goto("/welcome");
  }

  function handleChatAction(event) {
    showModal = false;
    sendMessage(event.detail.command);
  }

  async function handleApiAction(event) {
    isProcessingModal = true;
    if (event.detail.action === "create") {
      try {
        const res = await fetch("http://localhost:8080/api/accounts/create", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${$token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ type: event.detail.type }),
        });

        if (res.ok) {
          await fetchBalance();
          showNotification("Account created successfully", "success");
          showModal = false;
        } else {
          const data = await res.json();
          showNotification(data.error || "Error creating account", "error");
        }
      } catch (error) {
        showNotification("Server connection error", "error");
      }
    }
    isProcessingModal = false;
  }

  onMount(() => {
    window.history.pushState(null, null, window.location.href);
    const preventBack = () =>
      window.history.pushState(null, null, window.location.href);
    window.addEventListener("popstate", preventBack);
    fetchBalance();
    return () => window.removeEventListener("popstate", preventBack);
  });

  async function fetchBalance() {
    try {
      const res = await fetch("http://localhost:8080/api/balance", {
        method: "GET",
        headers: { Authorization: `Bearer ${$token}` },
      });

      if (res.ok) {
        const data = await res.json();
        accounts = data.accounts || [];
      }
    } catch (error) {
      console.error("Error getting balances:", error);
    }
  }

  async function fetchAccountHistory(account) {
    historyAccountLabel = account.account_number;
    showHistory = true;
    loadingHistory = true;
    historyTransactions = [];

    try {
      // Your Go endpoint receives the TBID by Query Param
      const res = await fetch(
        `http://localhost:8080/api/history?account_id=${account.tb_id}`,
        {
          method: "GET",
          headers: { Authorization: `Bearer ${$token}` },
        },
      );

      if (res.ok) {
        const data = await res.json();
        // Go returns { account_id, count, history: [...] }
        historyTransactions = data.history || [];
      } else {
        const err = await res.json();
        showNotification(err.error || "Could not load history", "error");
        showHistory = false;
      }
    } catch (error) {
      showNotification("Network error when consulting history", "error");
      showHistory = false;
    } finally {
      loadingHistory = false;
    }
  }

  // Helper function to keep the chat always at the bottom
  function scrollToBottom() {
    const chatBox = document.getElementById("chat-container");
    if (chatBox) {
      chatBox.scrollTo({ top: chatBox.scrollHeight, behavior: "smooth" });
    }
  }

  async function sendMessage(customMessage = null) {
    const userMsg = customMessage || message;
    if (!userMsg.trim() || loadingChat) return;

    message = "";
    chatHistory = [...chatHistory, { role: "user", content: userMsg }];
    loadingChat = true;

    await tick();
    scrollToBottom();

    try {
      const res = await fetch("http://localhost:8080/api/chat", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${$token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          message: userMsg,
          history: chatHistory.slice(-8).map((m) => ({
            role: m.role,
            content: m.content,
          })),
        }),
      });

      if (!res.ok) throw new Error(`HTTP Error: ${res.status}`);

      const data = await res.json();
      let botReply = data.reply || "";

      botReply = botReply
        .replace(/\(tool_call\)[\s\S]*?(\]|\)|\s|$)/g, "")
        .replace(/\(function=.*?\)/g, "")
        .replace(/<\/?[^>]+(>|$)/g, "")
        .trim();

      if (!botReply) botReply = "Operación completada. ✅ ¿Necesitas algo más?";

      chatHistory = [...chatHistory, { role: "assistant", content: botReply }];

  
      const lower = botReply.toLowerCase();
      const needsRefresh = ["✅", "exitoso", "depositado", "retirado", "transferencia", "realizado"].some(word => lower.includes(word));

      if (needsRefresh) {
      
        setTimeout(async () => {
          await fetchBalance();
          showNotification("Saldos actualizados en tiempo real", "success");
        }, 400);
      }
    } catch (error) {
      console.error("Chat error:", error);
      chatHistory = [...chatHistory, { role: "assistant", content: "❌ Error de conexión." }];
    } finally {
      loadingChat = false;
      await tick();
      scrollToBottom();
    }
  }

  async function deleteAccount(accountNumber) {
    try {
      const res = await fetch(
        `http://localhost:8080/api/accounts/${accountNumber}`,
        {
          method: "DELETE",
          headers: { Authorization: `Bearer ${$token}` },
        },
      );

      if (res.ok) {
        accounts = accounts.filter((a) => a.account_number !== accountNumber);
        showNotification(`Account ${accountNumber} deleted`, "success");
      } else {
        const data = await res.json();
        showNotification(data.error || "Error deleting account", "error");
      }
    } catch (err) {
      showNotification("Connection error", "error");
    }
  }

  async function handleManualTransaction(event) {
    const { action, data } = event.detail;
    if (isProcessingModal) return;

    isProcessingModal = true;

    let endpoint = "";
    let payload = {};

    if (action === "deposit") {
      endpoint = "/api/deposit";
      payload = {
        account_id: data.sourceID,
        amount: Math.round(parseFloat(data.amount) * 100),
      };
    } else if (action === "withdraw" || action === "transfer") {
      endpoint = "/api/transfer";
      payload = {
        from_account_id: data.sourceID,
        target_account_id: action === "withdraw" ? "1" : data.target,
        amount: Math.round(parseFloat(data.amount) * 100),
      };
    }

    try {
      const res = await fetch(`http://localhost:8080${endpoint}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${$token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        chatHistory = [
          ...chatHistory,
          {
            role: "assistant",
            content: `✅ ${action} operation carried out successfully.`,
          },
        ];
        await fetchBalance();
        showNotification("Transaction processed correctly", "success");
        showModal = false;
      } else {
        const errData = await res.json();
        const errorMsg =
          typeof errData.error === "string" ? errData.error : "Check the data";
        showNotification("Transaction rejected: " + errorMsg, "error");
      }
    } catch (err) {
      showNotification("Network error when processing transaction", "error");
    } finally {
      isProcessingModal = false;
    }
  }
</script>

{#if notification.show}
  <div
    transition:fly={{ y: -20, duration: 400 }}
    class="fixed top-6 right-6 z-[100] flex items-center gap-3 px-5 py-4 rounded-2xl shadow-2xl text-white font-bold max-w-sm
           {notification.type === 'error'
      ? 'bg-rose-600 shadow-rose-200'
      : 'bg-emerald-600 shadow-emerald-200'}"
  >
    {#if notification.type === "error"}
      <svg
        class="w-6 h-6 shrink-0"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        /></svg
      >
    {:else}
      <svg
        class="w-6 h-6 shrink-0"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
        /></svg
      >
    {/if}
    <span class="text-sm tracking-wide leading-tight flex-1"
      >{notification.message}</span
    >
    <button
      on:click={() => (notification.show = false)}
      class="ml-2 hover:opacity-75 transition-opacity"
      aria-label="Close notification"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M6 18L18 6M6 6l12 12"
        /></svg
      >
    </button>
  </div>
{/if}

<div class="min-h-screen bg-slate-50 flex flex-col font-sans text-slate-900">
  <header
    class="bg-white border-b border-slate-200 h-16 flex items-center justify-between px-8 sticky top-0 z-30 shadow-sm"
  >
    <div class="flex items-center gap-2">
      <div
        class="w-8 h-8 bg-blue-700 text-white flex items-center justify-center font-bold rounded-sm"
      >
        N
      </div>
      <span class="text-xl font-bold tracking-tight text-blue-700"
        >Nexora<span class="text-slate-500 font-light">Bank</span></span
      >
    </div>
    <div class="flex items-center gap-6">
      <span class="text-sm font-medium text-slate-600"
        >Hello, {displayName}</span
      >
      <button
        on:click={handleLogout}
        class="text-sm font-bold text-slate-400 hover:text-rose-600 transition-colors"
      >
        Log out
      </button>
    </div>
  </header>

  <div
    class="flex-1 max-w-7xl w-full mx-auto p-8 grid grid-cols-1 lg:grid-cols-12 gap-8"
  >
    <aside class="lg:col-span-3 space-y-2">
      <div
        class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-4 px-3"
      >
        Your Portal
      </div>
      <button
        class="w-full text-left px-4 py-3 bg-blue-50 text-blue-700 font-bold rounded-lg transition-colors"
        >Account Summary</button
      >
    </aside>

    <main class="lg:col-span-9 space-y-8">
      <div
        class="bg-white p-4 rounded-2xl shadow-sm border border-slate-200 flex flex-wrap gap-4"
      >
        <button
          on:click={() => {
            modalType = "crear";
            showModal = true;
          }}
          class="flex-1 min-w-[120px] p-4 rounded-xl hover:bg-blue-50 text-blue-700 font-bold transition-colors"
          >New Account</button
        >
        <button
          on:click={() => {
            modalType = "depositar";
            showModal = true;
          }}
          class="flex-1 min-w-[120px] p-4 rounded-xl hover:bg-emerald-50 text-emerald-700 font-bold transition-colors"
          >Deposit</button
        >
        <button
          on:click={() => {
            modalType = "retirar";
            showModal = true;
          }}
          class="flex-1 min-w-[120px] p-4 rounded-xl hover:bg-rose-50 text-rose-700 font-bold transition-colors"
          >Withdraw</button
        >
        <button
          on:click={() => {
            modalType = "transferir";
            showModal = true;
          }}
          class="flex-1 min-w-[120px] p-4 rounded-xl hover:bg-purple-50 text-purple-700 font-bold transition-colors"
          >Transfer</button
        >
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div class="space-y-6">
          <h2 class="text-xl font-extrabold tracking-tight">
            My Active Accounts
          </h2>

          {#if accounts.length === 0}
            <div
              class="p-10 border-2 border-dashed border-slate-200 rounded-3xl text-center bg-white shadow-sm flex flex-col items-center justify-center h-64"
            >
              <div
                class="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center mb-4 text-slate-300"
              >
                <svg
                  class="w-8 h-8"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                  /></svg
                >
              </div>
              <h3 class="text-slate-700 font-bold text-lg mb-1">
                No active accounts
              </h3>
              <p class="text-slate-500 text-sm">
                Create a "New Account" to start managing your funds in Nexora.
              </p>
            </div>
          {:else}
            <div class="space-y-4">
              {#each accounts as acc}
                <AccountCard
                  {acc}
                  on:delete={(e) => deleteAccount(e.detail)}
                  on:history={(e) => fetchAccountHistory(e.detail)}
                />
              {/each}
            </div>
          {/if}
        </div>

        <BankingPanel {accounts} on:refresh={fetchBalance} />
      </div>
    </main>
  </div>
</div>

{#if showModal}
  <TransactionModal
    {modalType}
    {accounts}
    isProcessing={isProcessingModal}
    on:close={() => (showModal = false)}
    on:apiAction={handleApiAction}
    on:manualTransaction={handleManualTransaction}
  />
{/if}

{#if showHistory}
  <HistoryModal
    show={showHistory}
    transactions={historyTransactions}
    accountName={historyAccountLabel}
    isLoading={loadingHistory}
    on:close={() => (showHistory = false)}
  />
{/if}
