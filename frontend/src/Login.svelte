<script>
  function login(realm) {
    const expiresAt = Date.now() + 3600 * 24 * 14;
    window.location.href = `https://api.worldoftanks.${realm}/wot/auth/login/?application_id=80aa6f7e1d5df049c79d9141dd0826ae&expires_at=${expiresAt}&redirect_uri=https://whaling-api.in.fkn.space/login?realm=${realm}`;
    // &redirect_uri=https://whaling.in.fkn.space/api/login`;
  }
</script>

<div class="w-full md:w-3/4 2xl:w-1/2 mx-auto">
  <div class="bg-gray-800 text-gray-200 rounded-sm text-md p-4">
    <div class="text-left text-gray-200">
      Welcome to your favorite Whaling website! On here, you'll be able to track
      your progress for the Snowflake event in World of Warships. This website
      will tell you how many resources you can earn, and even help you track
      your progress as you earn the event resources.
      <strong>
        Please note that the website may not work with hidden profiles.
      </strong>
    </div>

    <div
      class="mx-auto grid grid-cols-1 gap-2 mt-8 justify-center w-full md:w-1/2 lg:w-1/4 xl:w-1/6 px-8"
    >
      <button
        on:click={() => login('asia')}
        class="flex-grow font-medium px-5 py-2 bg-yellow-200 hover:bg-yellow-300 border-none text-yellow-900 rounded"
        >ASIA</button
      >
      <button
        on:click={() => login('ru')}
        class="font-medium px-5 py-2 bg-red-200 hover:bg-red-300 border-none text-red-900 rounded"
        >CIS</button
      >
      <button
        on:click={() => login('eu')}
        class="font-medium px-5 py-2 bg-blue-200 hover:bg-blue-300 border-none text-blue-900 rounded"
        >EU</button
      >
      <button
        on:click={() => login('com')}
        class="font-medium px-5 py-2 bg-purple-200 hover:bg-purple-300 border-none text-purple-900 rounded"
        >NA</button
      >
    </div>

    <div class="mt-16 text-left text-xs">
      <span class="block uppercase text-xs mb-2">Tracking information</span>
      Wargaming has changed the event rules so that you either need to win a battle
      or earn 300 base xp to blow off the snowflake. Due to limitations in the Wargaming
      API, the 300 base xp requirement cannot be tracked by this website. Since I
      hope that you are not a terrible player, I will assume that
      <strong>any battle</strong>
      will blow off a snowflake of your ship.
    </div>
  </div>
</div>
