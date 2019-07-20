<script>
    function login() {
        const expiresAt = Date.now() + 3600 * 24 * 14;
        window.location.href = `https://api.worldoftanks.eu/wot/auth/login/
            ?application_id=80aa6f7e1d5df049c79d9141dd0826ae
            &expires_at=${expiresAt}
            &redirect_uri=https://vh66uhz6ce.execute-api.eu-central-1.amazonaws.com/dev/login`;
            // &redirect_uri=https://frenchwhaling.in.fkn.space/api/login`;
    }
</script>
<div class="w-3/4 text-center">
    <div class="bg-gray-200 text-gray-600 rounded-sm text-md p-4">
        To figure out how much of a whale you really are, we'll need you to log in with the Wargaming website.
    </div>

    <button on:click={login} class="mt-8 px-5 py-2 bg-red-200 hover:bg-red-300 border-none text-red-900 rounded">Login</button>
</div>
