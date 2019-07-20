<script>
    import {onMount} from 'svelte';
    import Login from './Login.svelte';
    import * as querystring from 'query-string';
    import jwtDecode from 'jwt-decode';
    import {accountId, token, nickname, loggedIn} from './store';

    let toggle = false;
    let error = false;

    onMount(() => {
        const data = querystring.parseUrl(window.location.href);

        if (data.query && data.query.success === 'true') {
            try {
                $loggedIn = true;
                const parsed = jwtDecode(data.query.token);

                $accountId = parsed.sub;
                $nickname = parsed.nickname;
                $token = data.query.token;
            } catch(e) {
                error = true;
                $loggedIn = false;
            }
        } else if (data.query && data.query.success === 'false') {
            $loggedIn = false;
            error = true;
        }
    });

    function logout() {
        $loggedIn = false;
        $token = undefined;
    }
</script>

<div class="font-sans w-full h-screen">
    <div class="w-1/5 invisible md:visible bg-blue-100 md:float-left h-0 md:h-screen"></div>
    <div class="w-1/5 invisible md:visible bg-red-100 shadow-inner md:float-right h-0 md:h-screen"></div>
    <div class="w-auto h-screen bg-white shadow-md overflow-x-hidden">
        <div class="flex flex-wrap mt-4 p-4">
            <div class="w-5/5 md:w-2/5 mx-auto h-24 md:h-80">
                <img alt="Logo made by AdonisWerther" class="h-24 md:h-80 w-auto mx-auto" src="/img/hon.png" />
            </div>
            <div class="w-5/5 md:w-3/5 flex-grow">
                <h1 class="text-5xl">Frenchwhaling</h1>
                <div class="-mt-2 text-gray-600">Brought to you by the same idiot who did Steelwhaling</div>
            </div>
        </div>

        {#if $loggedIn}
            <button on:click={logout} class="mr-4 float-right px-4 font-xs border-none py-1 rounded bg-gray-200 hover:bg-gray-400 text-gray-700">Logout</button>
        {/if}

        <div class="flex flex-wrap mt-32 justify-around">
            {#if error}
            <div class="w-full flex justify-around mb-8">
                <div class="w-1/2 text-center bg-red-600 text-white rounded-sm p-4">
                    There was an error logging you in.
                </div>
            </div>
            {/if}

            {#if $loggedIn}
                Hi there, {$nickname}({$accountId}).
            {:else}
                <Login />
            {/if}
        </div>
        <div class="mt-8 text-gray-600 text-sm text-center"><a href="#privacy" on:click={() => toggle = !toggle}>Privacy Policy</a></div>

        {#if toggle}
        <div class="mt-8 pl-8 text-md text-gray-600 text-left">
            <p>To provide this service for you, the following data will be collected and stored:</p>

            <ul class="pl-8 mt-2">
                <li>Your Wargaming user id</li>
                <li>Statistics about battles you played in World of Warships</li>
            </ul>

            <p class="mt-2 ">The data will be stored on Amazon Web Services in the eu-central-1 (Frankfurt) region. I intend to publish statistics about the event after it is over. Data will be completely anonymised and only general statistics will be provided, which will not be traceable back to an individual player.</p>
        </div>
        {/if}
    </div>
</div>