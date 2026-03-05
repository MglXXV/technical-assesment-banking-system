<script>
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';

  const dispatch = createEventDispatcher();

  export let show = false;
  export let transactions = []; // Recibe solo las 5 de la página actual
  export let metadata = { current_page: 1, has_more: false };
  export let isLoading = false;
  export let accountName = "";


  const formatMoney = (amountStr) => {
    const val = parseFloat(amountStr) / 100;
    return val.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
  };

  const formatDate = (dateStr) => {
    const d = new Date(dateStr);
    return new Intl.DateTimeFormat('es-ES', { 
      day: '2-digit', month: 'short', year: 'numeric', 
      hour: '2-digit', minute: '2-digit' 
    }).format(d);
  };

  // Función para disparar el cambio de página
  function goToPage(step) {
    const newPage = metadata.current_page + step;
    dispatch('changePage', newPage);
  }
</script>

{#if show}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[60] flex items-center justify-center p-4" 
    on:click|self={() => dispatch('close')} 
    role="button"
    tabindex="-1"
    transition:fade>
    
    <div class="bg-white rounded-3xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[85vh]" in:fly={{ y: 20 }}>
      
      <div class="p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50 shrink-0">
        <div>
          <h3 class="text-xl font-extrabold text-slate-800">Recent Movements</h3>
          <p class="text-xs font-bold text-slate-500 uppercase tracking-widest mt-1">Account: {accountName}</p>
        </div>
        <!-- svelte-ignore a11y_consider_explicit_label -->
        <button on:click={() => dispatch('close')} class="p-2 bg-white rounded-full shadow-sm text-slate-400">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <div class="p-6 overflow-y-auto flex-1 bg-slate-50/50">
        {#if isLoading}
          <div class="flex flex-col items-center justify-center h-40">
             <div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full mb-4"></div>
             <span class="text-sm font-bold text-slate-500">Querying TigerBeetle...</span>
          </div>
        {:else if transactions.length === 0}
          <div class="text-center py-10">
            <p class="text-slate-500">No movements found.</p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each transactions as tx}
              <div class="bg-white p-4 rounded-2xl border border-slate-100 shadow-sm flex items-center justify-between">
                <div class="flex items-center gap-4">
                  <div class="w-10 h-10 rounded-xl flex items-center justify-center {tx.type === 'CREDIT' ? 'bg-emerald-50 text-emerald-600' : 'bg-rose-50 text-rose-600'}">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path d={tx.type === 'CREDIT' ? "M19 14l-7 7m0 0l-7-7m7 7V3" : "M5 10l7-7m0 0l7 7m-7-7v18"} />
                    </svg>
                  </div>
                  <div>
                    <p class="font-bold text-slate-800 text-sm">{tx.type === 'CREDIT' ? 'Credit Received' : 'Debit Sent'}</p>
                    <p class="text-[10px] text-slate-400">{formatDate(tx.date)}</p>
                  </div>
                </div>
                <p class="font-bold {tx.type === 'CREDIT' ? 'text-emerald-600' : 'text-slate-800'}">
                  {tx.type === 'CREDIT' ? '+' : '-'}{formatMoney(tx.amount)}
                </p>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <div class="p-4 border-t border-slate-100 bg-white flex justify-between items-center shrink-0">
        <button 
          on:click={() => goToPage(-1)}
          disabled={metadata.current_page <= 1 || isLoading}
          class="px-4 py-2 text-xs font-bold text-slate-600 hover:bg-slate-100 rounded-xl disabled:opacity-30 transition-all"
        >
          ← Previous
        </button>

        <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">
          Page {metadata.current_page}
        </span>

        <button 
          on:click={() => goToPage(1)}
          disabled={!metadata.has_more || isLoading}
          class="px-4 py-2 text-xs font-bold text-blue-700 hover:bg-blue-50 rounded-xl disabled:opacity-30 transition-all"
        >
          Next →
        </button>
      </div>
    </div>
  </div>
{/if}