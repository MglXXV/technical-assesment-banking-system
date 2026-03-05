<script>
  import { createEventDispatcher } from "svelte";
  export let acc;
  const dispatch = createEventDispatcher();

  const typeLabels = {
    savings: "Savings Account",
    checking: "Checking Account",
    investment: "Investment Account",
  };

  // Format the balance to ensure 2 decimal places (UX: Banking consistency)
  $: formattedBalance = typeof acc.balance === 'number' 
    ? acc.balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    : "0.00";

  function handleDelete() {
    if (confirm(`Are you sure you want to delete the account ${acc.account_number}? This action is irreversible.`)) {
      dispatch("delete", acc.account_number);
    }
  }

  // UX: State to give visual feedback when copying without using annoying alerts
  let copied = false;
  function copyNumber() {
    navigator.clipboard.writeText(acc.account_number);
    copied = true;
    setTimeout(() => { copied = false; }, 2000); // The green check lasts 2 seconds
  }

  function viewHistory() {
    dispatch("history", acc);
  }
</script>

<div class="relative bg-gradient-to-br from-slate-900 to-slate-800 p-8 rounded-3xl text-white shadow-xl group overflow-hidden transition-all hover:shadow-2xl hover:-translate-y-1 border border-white/5">
  
  <div class="absolute -right-10 -top-10 w-40 h-40 bg-blue-500/10 rounded-full blur-3xl transition-colors group-hover:bg-blue-500/20"></div>

  <button
    on:click={handleDelete}
    class="absolute top-6 right-6 text-slate-500 hover:text-rose-400 transition-colors z-20"
    title="Close account"
  >
    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
    </svg>
  </button>

  <div class="relative z-10 flex flex-col h-full justify-between gap-6">
    
    <div class="flex justify-between items-start">
      <span class="text-[10px] font-bold uppercase tracking-widest text-blue-300 bg-blue-900/50 px-3 py-1 rounded-full border border-blue-500/30 backdrop-blur-sm">
        {typeLabels[acc.type] || acc.type}
      </span>
    </div>

    <div>
      <p class="text-sm text-slate-400 mb-1">Available Balance</p>
      <div class="flex items-baseline gap-2">
        <h3 class="text-4xl font-light font-mono text-white">
          ${formattedBalance}
        </h3>
        <span class="text-sm font-bold opacity-50 uppercase">{acc.currency || 'USD'}</span>
      </div>
    </div>

    <div class="flex justify-between items-end mt-2 pt-4 border-t border-white/10">
      
      <div class="flex items-center gap-2">
        <span class="text-sm font-mono tracking-widest text-slate-300">{acc.account_number}</span>
        
        <button on:click={copyNumber} class="hover:text-white transition-colors" title="Copy account number">
          {#if copied}
            <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
            </svg>
          {:else}
            <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
            </svg>
          {/if}
        </button>
      </div>

      <button 
        on:click={viewHistory}
        class="text-xs font-bold flex items-center gap-1.5 text-blue-400 hover:text-blue-300 transition-colors bg-blue-500/10 hover:bg-blue-500/20 px-3 py-2 rounded-lg"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        History
      </button>

    </div>
  </div>
</div>