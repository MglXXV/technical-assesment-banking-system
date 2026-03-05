<script>
    import { Route as TinroRoute, router } from 'tinro';
    const Route = /** @type {any} */ (TinroRoute);
    import { token } from './lib/authStore';
    
    import Welcome from './lib/views/Welcome.svelte';
    import Auth from './lib/views/Auth.svelte';
    import Dashboard from './lib/views/Dashboard.svelte';

    // Smart Route Protection and Redirection
    router.subscribe(page => {
        // 1. Logic if the user is NOT logged in
        if (!$token) {
            // If the user enters the root "/", we automatically send them to "/welcome"
            if (page.path === '/') {
                router.goto('/welcome', true);
            }
            // If the user tries to sneak into the dashboard, we send them to "/login"
            else if (page.path === '/dashboard') {
                router.goto('/login', true);
            }
        } 
        // 2. Logic if the user IS logged in
        else {
            // If the user tries to go back to the root, welcome or login, we send them back to the dashboard
            if (['/', '/welcome', '/login', '/register'].includes(page.path)) {
                router.goto('/dashboard', true);
            }
        }
    });
</script>

<Route path="/welcome"><Welcome /></Route>

<Route path="/login"><Auth mode="login" /></Route>
<Route path="/register"><Auth mode="register" /></Route>

<Route path="/dashboard"><Dashboard /></Route>