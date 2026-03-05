<script>
  import { createEventDispatcher } from "svelte";
  import { fade, fly } from 'svelte/transition';

  const dispatch = createEventDispatcher();

  export let modalType; // 'deposit', 'withdraw', 'transfer', 'create'
  export let accounts = [];
  export let isProcessing = false; // LOADING STATE

  let amount = "";
  let targetAccount = "";
  let txAccType = "savings";
  let sourceAccount = "";

  // Reactivity: Select the first account by default when they load
  $: if (accounts.length > 0 && !sourceAccount) {
    sourceAccount = accounts[0].tb_id;
  }

  function handleSelectChange(e) {
    sourceAccount = e.target.value;
  }

  function handleSubmit() {
    if (modalType === "crear") {
      dispatch("apiAction", { action: "create", type: txAccType });
      return;
    }

    // Last ID validation defense
    const finalSource = sourceAccount || (accounts.length > 0 ? accounts[0].tb_id : undefined);

    let actionName = "transfer";
    if (modalType === "depositar") actionName = "deposit";
    if (modalType === "retirar") actionName = "withdraw";

    dispatch("manualTransaction", {
      action: actionName,
      data: {
        amount: amount,
        sourceID: finalSource,
        target: targetAccount,
      },
    });
  }
</script>

<div class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-50 flex items-center justify-center p-4" 
  on:click|self={() => !isProcessing && dispatch('close')} 
  on:keydown|self={(e) => !isProcessing && (e.key === 'Enter' || e.key === ' ') && dispatch('close')}
  role="button"
  tabindex="-1"
  transition:fade>
  
  <div class="bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden" in:fly={{ y: 20 }}>
    <div class="p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50">
      <h3 class="text-xl font-bold text-slate-800 capitalize">{modalType}</h3>
      <button on:click={() => !isProcessing && dispatch('close')} disabled={isProcessing} class="text-slate-400 hover:text-slate-600 disabled:opacity-50" aria-label="Close">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
      </button>
    </div>

    <form on:submit|preventDefault={handleSubmit} class="p-8 space-y-6">
      
      {#if modalType === 'crear'}
        <div>
          <label for="txAccType" class="block text-xs font-bold text-slate-500 uppercase mb-2">Account Type</label>
          <select id="txAccType" bind:value={txAccType} disabled={isProcessing} class="w-full px-4 py-3 bg-slate-50 border rounded-xl outline-none focus:ring-2 focus:ring-blue-600 disabled:opacity-60">
            <option value="savings">Savings</option>
            <option value="checking">Checking</option>
            <option value="investment">Investment</option>
          </select>
        </div>
      {:else}
        <div>
          <label for="amount" class="block text-xs font-bold text-slate-500 uppercase mb-2">Amount (USD)</label>
          <input id="amount" type="number" step="0.01" min="0.01" bind:value={amount} required disabled={isProcessing} placeholder="0.00" class="w-full px-4 py-3 bg-slate-50 border rounded-xl text-2xl font-mono outline-none focus:ring-2 focus:ring-blue-600 disabled:opacity-60" />
        </div>

        <div>
          <label for="sourceAccount" class="block text-xs font-bold text-slate-500 uppercase mb-2">Source / Destination Account</label>
          <select id="sourceAccount" value={sourceAccount} on:change={handleSelectChange} disabled={isProcessing} required class="w-full px-4 py-3 bg-slate-50 border rounded-xl outline-none focus:ring-2 focus:ring-blue-600 disabled:opacity-60">
            {#each accounts as acc}
              <option value={acc.tb_id}>{acc.account_number} ({acc.type}) - ${acc.balance || '0.00'}</option>
            {/each}
          </select>
        </div>

        {#if modalType === 'transferir'}
          <div>
            <label for="targetAccount" class="block text-xs font-bold text-slate-500 uppercase mb-2">Destination Account Number</label>
            <input id="targetAccount" type="text" bind:value={targetAccount} required disabled={isProcessing} placeholder="Ex: 4001-0001-1000" class="w-full px-4 py-3 bg-slate-50 border rounded-xl font-mono outline-none focus:ring-2 focus:ring-blue-600 disabled:opacity-60" />
          </div>
        {/if}
      {/if}

      <button type="submit" disabled={isProcessing} class="w-full py-4 bg-blue-700 text-white font-bold rounded-xl hover:bg-blue-800 transition-all shadow-lg flex justify-center items-center disabled:opacity-75 disabled:cursor-not-allowed">
        {#if isProcessing}
          <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
          Processing...
        {:else}
          Confirm Operation
        {/if}
      </button>
    </form>
  </div>
</div>