<script>
    import {onMount} from 'svelte';
    import Login from './Login.svelte';
    import Progress from './Progress.svelte';
    import * as querystring from 'query-string';
    import jwtDecode from 'jwt-decode';
    import {accountId, dataUrl, token, nickname, realm, loggedIn} from './store';
    import {reportClick} from './clickEvents';

    let toggle = false;
    let error = false;
    let reason = 'UNKNOWN';
    let isNew = false;

    const eventStartTimes = {
        eu:   1564034400,
        com:  1563963722,
        ru:   1563937200,
        asia: 1564005600,
    };
    const ts = Math.round(+ new Date() / 1000);

    onMount(() => {
        const data = querystring.parseUrl(window.location.href);
        isNew = data.query.isNew === 'true';

        if (data.query && data.query.success === 'true') {
            try {
                $loggedIn = true;
                const parsed = jwtDecode(data.query.token);

                $accountId = parsed.sub;
                $nickname = parsed.nickname;
                $realm = parsed.realm;
                $token = data.query.token;
                $dataUrl = data.query.dataUrl;
            } catch(e) {
                console.log(e);
                error = true;
                $loggedIn = false;
            }
        } else if (data.query && data.query.success === 'false') {
            $loggedIn = false;
            reason = data.query.reason;
            error = true;
        }
    });

    function logout() {
        $loggedIn = false;
        $token = undefined;
        window.history.pushState('', 'WoWS Whaling', '/');

        reportClick('Logout');
    }

    function donate() {
        reportClick('Donate');
        alert('Thanks for clicking on this button! This was a project I built during my free time and I am paying the infrastructure costs myself. While I do not take actual money as donations, I am always happy to read a "Thank you" email or receiving a little gift on EU, my username is Rukenshia.');
    }

    function contact() {
        reportClick('Contact');
    }

    function privacyPolicy() {
        toggle = !toggle;

        if (toggle) {
            reportClick('PrivacyPolicy');
        }
    }
</script>

<style>
a:visited {
    @apply text-gray-400;
}

body {
    @apply text-white;
}
</style>

<div class="font-sans w-full h-screen text-white">
    <div class="w-1/12 invisible md:visible bg-red-900 md:float-left h-0 md:h-screen"></div>
    <div class="w-1/12 invisible md:visible bg-red-900 shadow-inner md:float-right h-0 md:h-screen"></div>
    <div class="w-auto h-screen bg-gray-900 shadow-md overflow-x-hidden overflow-y-visible">
        <div class="flex flex-wrap mt-4 p-4">
            <div class="w-5/5 md:w-2/5 mx-auto h-24 md:h-80">
                <img alt="Logo made by AdonisWerther" class="h-24 md:h-80 w-auto float-right" src="/img/hon.png" />
            </div>
            <div class="w-5/5 pl-4 md:w-3/5 flex-grow">
                <h1 class="text-5xl text-gray-300">Steelwhaling</h1>
                <div class="-mt-2 text-gray-500">Brought to you by Rukenshia on the EU server, the same idiot that built Steelwhaling, Frenchwhaling and Shipvoting</div>
            </div>
        </div>

        {#if $loggedIn}
            <div class="mt-12 w-full flex justify-around">
                <div class="w-full xl:w-3/4">
                    <div class="float-right h-8">
                        <button on:click={donate} class="mr-4 px-4 font-xs border-none py-1 rounded bg-green-400 hover:bg-green-500 text-gray-800 shadow-md">Donate</button>
                        <a on:click={contact} style="padding-top: 7px; padding-bottom: 7px;" href="mailto:svc-frenchwhaling@ruken.pw" class="mr-4 p-0 px-4 border-none rounded bg-gray-700 hover:no-underline hover:bg-gray-800">Contact me</a>
                        <button on:click={logout} class="mr-4 px-4 font-xs border-none py-1 rounded bg-gray-700 hover:bg-gray-800">Logout</button>
                    </div>
                    {#if ts < eventStartTimes[$realm]}
                        <div class="w-full flex justify-around mt-16 mb-8">
                            <div class="w-3/4 bg-blue-300 text-blue-900 font-medium rounded p-4">
                                You are preregistered for the tracking, but the event has not started on your server yet. Data will update as soon
                                as the patch is live on your server and you started playing battles.
                            </div>
                        </div>
                    {/if}
                    <Progress {isNew} />
                </div>
            </div>
        {:else}    
            <div class="mt-16 w-full flex justify-around">
                <div class="w-2/3 text-gray-500">
                    Welcome to your favorite Whaling website! On here, you'll be able to track your progress for various
                    World of Warships events such as the Snowflake 2019 event.
                    <strong>Please note that the website may not work with hidden profiles.</strong>
                </div>
            </div>
            <div class="flex flex-wrap mt-8 justify-around">
                {#if error}
                <div class="w-full flex justify-around mb-8">
                    <div class="w-1/2 text-center bg-red-600 text-white rounded-sm p-4">
                        There was an error logging you in. Feel free to contact me <a class="font-medium underline" href="mailto:svc-frenchwhaling@ruken.pw">via Email</a> or Discord (Rukenshia#4396) if you can't get past this.
                        <br />
                        Error message: <span class="font-mono">{reason}<span>
                    </div>
                </div>
                {/if}

                <Login />
            </div>
            <div class="mb-32"></div>
        {/if}
        <div class="mt-8 mb-8 text-gray-400 font-medium text-sm text-center">
            <a href="#privacy" on:click={privacyPolicy}>Privacy Policy</a>
            &bullet;
            <a target="_blank" href="https://git.sr.ht/~rukenshia/frenchwhaling">Source code</a>
            &bullet;
            This website is not affiliated with Wargaming
            &bullet;
            Thanks to AdonisWerther for the logo ❤️
        </div>

        {#if toggle}
        <div class="flex justify-around">
            <div id="privacy" class="w-3/4 mb-8 pl-8 text-md text-gray-400 text-left">
                <h3 class="text-3xl mb-4">Privacy Policy</h3>
                <p>To provide this service to you, the following data will be collected and stored:</p>

                <ul class="pl-8 mt-2">
                    <li>Your Wargaming account id</li>
                    <li>Statistics about ships you own in World of Warships</li>
                    <li>Statistics about how you perform in battles (wins per game mode)</li>
                </ul>

                <p class="mt-2 ">The data will be stored on Amazon Web Services in the eu-central-1 (Frankfurt) region. I intend to publish statistics about the event after it is over. Data will be anonymised and only general statistics will be provided to the public, meaning that published data will not be traceable to individuals.</p>
            </div>
        </div>
        {/if}
    </div>
</div>