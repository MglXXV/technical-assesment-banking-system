<script>
    import { slide, fade, fly } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';
    import { token, user } from '../authStore';
    import { router } from 'tinro'; // We use the router
    
    export let mode = "login"; // We receive from the URL if it is login or register
    $: isLogin = mode === "login"; 
    
    let loading = false;
    let error = "";
    let successMsg = "";
    let email = "";
    let password = "";
    let fullname = "";

    async function handleSubmit() {
        loading = true;
        error = "";
        successMsg = "";
        const endpoint = isLogin ? '/api/login' : '/api/register';
        const payload = isLogin 
            ? { email, password } 
            : { name: fullname, email, password };

        try {
            const res = await fetch(`http://localhost:8080${endpoint}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || "An error occurred during authentication");

            if (isLogin) {
                successMsg = "Authentication successful. Loading your portal...";
                setTimeout(() => {
                    token.set(data.token);
                    user.set(data.user);
                    router.goto('/dashboard'); // WE TRAVEL TO THE DASHBOARD
                }, 1500);
            } else {
                successMsg = "Account created successfully. You can now log in.";
                setTimeout(() => {
                    successMsg = "";
                    password = ""; 
                    router.goto('/login'); // WE TRAVEL TO THE LOGIN
                }, 2000);
            }
        } catch (err) {
            error = err.message;
        } finally {
            if (!successMsg) loading = false; 
        }
    }
</script>

<div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
    <div class="bg-white rounded-2xl shadow-2xl overflow-hidden border border-slate-100 w-full max-w-md relative">
        
        {#if successMsg}
            <div transition:fly={{ y: -20, duration: 500, easing: quintOut }} 
                 class="absolute top-6 left-1/2 -translate-x-1/2 z-50 bg-emerald-50 border border-emerald-200 text-emerald-700 px-6 py-3 rounded-full shadow-lg flex items-center gap-3 font-semibold text-sm w-max max-w-[90%] text-center">
                <svg class="w-5 h-5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                {successMsg}
            </div>
        {/if}

        <a href="/welcome" aria-label="Go back" class="absolute top-4 right-4 w-8 h-8 flex items-center justify-center rounded-full bg-slate-100 text-slate-500 hover:bg-slate-200 hover:text-slate-800 transition-colors z-10">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </a>

        <div class="bg-slate-50 p-8 text-center border-b border-slate-100 mt-2">
            {#key isLogin}
                <div in:fade={{ duration: 400, delay: 150 }}>
                    <h2 class="text-2xl font-extrabold text-blue-700 tracking-tight">
                        {isLogin ? 'Secure Access' : 'Open your Account'}
                    </h2>
                    <p class="text-sm text-slate-500 mt-2 font-medium">
                        {isLogin ? 'Enter your Nexora Bank portal' : 'Join the new era of digital banking'}
                    </p>
                </div>
            {/key}
        </div>

        <div class="p-8">
            {#if error}
                <div transition:slide|local class="mb-6 p-3 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-semibold rounded-r-md">
                    <div in:fade>{error}</div>
                </div>
            {/if}

            <form on:submit|preventDefault={handleSubmit} class="space-y-5">
                {#if !isLogin}
                    <div transition:slide|local>
                        <div in:fade>
                            <label for="fullname" class="block text-xs font-bold text-slate-500 uppercase tracking-wide mb-1.5">Full Name</label>
                            <input id="fullname" bind:value={fullname} type="text" required disabled={loading || !!successMsg} class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600 outline-none transition-all text-slate-700 disabled:opacity-50" />
                        </div>
                    </div>
                {/if}

                <div>
                    <label for="email" class="block text-xs font-bold text-slate-500 uppercase tracking-wide mb-1.5">Email</label>
                    <input id="email" bind:value={email} type="email" required disabled={loading || !!successMsg} class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600 outline-none transition-all text-slate-700 disabled:opacity-50" />
                </div>

                <div>
                    <label for="password" class="block text-xs font-bold text-slate-500 uppercase tracking-wide mb-1.5">Password</label>
                    <input id="password" bind:value={password} type="password" required disabled={loading || !!successMsg} class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600 outline-none transition-all text-slate-700 disabled:opacity-50" />
                </div>

                <button type="submit" disabled={loading || !!successMsg} class="w-full py-4 bg-blue-700 text-white font-bold rounded-lg hover:bg-blue-800 transition-all shadow-lg shadow-blue-200 disabled:opacity-70">
                    {#if !!successMsg} Processing... {:else if loading} Connecting... {:else} {isLogin ? 'Enter' : 'Create Account'} {/if}
                </button>
            </form>

            <div class="mt-8 text-center">
                <a href={isLogin ? "/register" : "/login"} class="text-sm font-bold text-blue-700 hover:text-blue-800 hover:underline transition-colors">
                    {isLogin ? 'Don\'t have an account? Register here' : 'Already have an account? Log in'}
                </a>
            </div>
        </div>
    </div>
</div>