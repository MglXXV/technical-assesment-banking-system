<script>
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';

  const dispatch = createEventDispatcher();

  export let show = false;
  export let transactions = [];
  export let accountName = "";
  export let isLoading = false;

  // Currency formatter
  const formatMoney = (amountStr) => {
    // TigerBeetle returns the amount in cents as a string, e.g.: "10000" for $100.00
    const val = parseFloat(amountStr) / 100;
    return val.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
  };

  // Date formatter
  const formatDate = (dateStr) => {
    const d = new Date(dateStr);
    return new Intl.DateTimeFormat('es-ES', { 
      day: '2-digit', month: 'short', year: 'numeric', 
      hour: '2-digit', minute: '2-digit' 
    }).format(d);
  };
</script>

{#if show}
  <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[60] flex items-center justify-center p-4" 
    on:click|self={() => dispatch('close')} 
    on:keydown|self={(e) => (e.key === 'Enter' || e.key === ' ') && dispatch('close')}
    role="button"
    tabindex="-1"
    transition:fade>
    
    <div class="bg-white rounded-3xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[85vh]" in:fly={{ y: 20 }}>
      
      <div class="p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50 shrink-0">
        <div>
          <h3 class="text-xl font-extrabold text-slate-800">Recent Movements</h3>
          <p class="text-xs font-bold text-slate-500 uppercase tracking-widest mt-1">Account: {accountName}</p>
        </div>
        <button on:click={() => dispatch('close')} aria-label="Close" class="p-2 bg-white rounded-full shadow-sm text-slate-400 hover:text-slate-700 transition-colors">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <div class="p-6 overflow-y-auto flex-1 bg-slate-50/50">
        {#if isLoading}
          <div class="flex flex-col items-center justify-center h-40 gap-4">
            <svg class="animate-spin h-8 w-8 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
            <span class="text-sm font-bold text-slate-500">Loading history from TigerBeetle...</span>
          </div>
        {:else if transactions.length === 0}
          <div class="flex flex-col items-center justify-center h-40 text-center">
            <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mb-3 text-slate-300">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
            </div>
            <p class="text-slate-500 font-medium">No movements registered in this account.</p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each transactions as tx}
              <div class="bg-white p-4 rounded-2xl border border-slate-100 shadow-sm flex items-center justify-between hover:shadow-md transition-shadow">
                <div class="flex items-center gap-4">
                  <div class="w-12 h-12 rounded-xl flex items-center justify-center {tx.type === 'CREDIT' ? 'bg-emerald-50 text-emerald-600' : 'bg-rose-50 text-rose-600'}">
                    {#if tx.type === 'CREDIT'}
                      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" /></svg>
                    {:else}
                      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" /></svg>
                    {/if}
                  </div>
                  <div>
                    <p class="font-bold text-slate-800">{tx.type === 'CREDIT' ? 'Income received' : 'Transfer sent'}</p>
                    <p class="text-xs text-slate-400 font-medium mt-0.5">{formatDate(tx.date)} • Ref: {tx.transfer_id.substring(0,8)}...</p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="font-extrabold text-lg {tx.type === 'CREDIT' ? 'text-emerald-600' : 'text-slate-800'}">
                    {tx.type === 'CREDIT' ? '+' : '-'}{formatMoney(tx.amount)}
                  </p>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}