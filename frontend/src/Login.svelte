<script>
    function login(realm) {
        const expiresAt = Date.now() + 3600 * 24 * 14;
        window.location.href = `https://api.worldoftanks.${realm}/wot/auth/login/?application_id=80aa6f7e1d5df049c79d9141dd0826ae&expires_at=${expiresAt}&redirect_uri=https://whaling-api.in.fkn.space/login?realm=${realm}`;
        // &redirect_uri=https://whaling.in.fkn.space/api/login`;
    }
</script>

<div class="w-3/4 text-center">
    <div class="w-3/4 mx-auto">
        <div class="bg-gray-800 text-gray-400 rounded-sm text-md p-4">
            Please use the buttons below to log into your server.
        </div>
    </div>

    <div class="flex flex-wrap mt-8">
        <div class="w-1/4 p-4">
            <button
                on:click={() => login('asia')}
                class="font-medium px-5 py-2 bg-yellow-200 hover:bg-yellow-300 border-none text-yellow-900 rounded">ASIA</button>
        </div>
        <div class="w-1/4 p-4">
            <button
                on:click={() => login('ru')}
                class="font-medium px-5 py-2 bg-red-200 hover:bg-red-300 border-none text-red-900 rounded">CIS</button>
        </div>
        <div class="w-1/4 p-4">
            <button
                on:click={() => login('eu')}
                class="font-medium px-5 py-2 bg-blue-200 hover:bg-blue-300 border-none text-blue-900 rounded">EU</button>
        </div>
        <div class="w-1/4 p-4">
            <button
                on:click={() => login('com')}
                class="font-medium px-5 py-2 bg-purple-200 hover:bg-purple-300 border-none text-purple-900 rounded">NA</button>
        </div>
    </div>
</div>
