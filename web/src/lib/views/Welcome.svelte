<script>
    // We import router from tinro in case we need programmatic navigation
    import { router } from 'tinro';
</script>

<div class="min-h-screen bg-slate-50 font-sans text-slate-900">
    <nav class="fixed top-0 w-full bg-white/90 backdrop-blur-md border-b border-slate-200 z-40">
        <div class="max-w-7xl mx-auto px-6 h-20 flex justify-between items-center">
            <a href="/" class="flex items-center gap-2 cursor-pointer">
                <div class="w-8 h-8 bg-blue-700 text-white flex items-center justify-center font-bold text-xl rounded-sm">N</div>
                <span class="text-2xl font-bold tracking-tight text-blue-700">Nexora<span class="font-light text-slate-500">Bank</span></span>
            </a>
            
            <div class="hidden md:flex gap-8 text-sm font-semibold text-slate-600">
                <button class="hover:text-blue-700 transition-colors">Credit Cards</button>
                <button class="hover:text-blue-700 transition-colors">Savings Accounts</button>
                <button class="hover:text-blue-700 transition-colors">Loans</button>
            </div>

            <a href="/login" class="bg-blue-700 text-white px-6 py-2.5 font-semibold text-sm rounded-sm hover:bg-blue-800 transition-all shadow-md hover:shadow-lg">
                Log in
            </a>
        </div>
    </nav>

    <main class="relative pt-32 pb-16 lg:pt-48 lg:pb-32 overflow-hidden">
        <div class="max-w-7xl mx-auto px-6 grid lg:grid-cols-2 gap-16 items-center">
            <div class="space-y-8 relative z-10">
                <h1 class="text-5xl lg:text-7xl font-extrabold text-slate-900 leading-[1.1] tracking-tight">
                    Manage your money with <span class="text-blue-700">precision.</span>
                </h1>
                <p class="text-lg text-slate-600 max-w-lg leading-relaxed font-light">
                    High-performance accounts backed by the speed of TigerBeetle. Designed for those who demand excellence in digital banking.
                </p>
                <div class="flex flex-col sm:flex-row gap-4 pt-4">
                    <a href="/register" class="bg-blue-700 text-white px-8 py-4 font-bold rounded-sm shadow-xl shadow-blue-700/20 hover:bg-blue-800 transition-all text-center">
                        Open Account Now
                    </a>
                    <button class="bg-white text-blue-700 border border-slate-300 px-8 py-4 font-bold rounded-sm hover:bg-slate-50 transition-all">
                        Learn about Benefits
                    </button>
                </div>
            </div>

            <div class="relative z-10 hidden lg:block">
                <div class="absolute inset-0 bg-blue-100 transform translate-x-4 translate-y-4 rounded-xl -z-10"></div>
                <img src="https://images.unsplash.com/photo-1573164713988-8665fc963095?auto=format&fit=crop&q=80&w=1000" alt="Premium Experience" class="rounded-xl shadow-2xl object-cover h-[500px] w-full border border-slate-200" />
            </div>
        </div>
    </main>
</div>