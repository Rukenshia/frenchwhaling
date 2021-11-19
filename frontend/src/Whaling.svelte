<script>
  import { onMount } from 'svelte';
  import Login from './Login.svelte';
  import Modal from './Modal.svelte';
  import Progress from './Progress.svelte';
  import * as querystring from 'query-string';
  import HRNumbers from 'human-readable-numbers';
  import jwtDecode from 'jwt-decode';
  import { formatRelative } from 'date-fns';
  import {
    accountId,
    dataUrl,
    token,
    nickname,
    realm,
    loggedIn,
    statistics,
    resourceName,
  } from './store';
  import { reportClick } from './clickEvents';

  let toggle = false;
  let error = false;
  let reason = 'UNKNOWN';
  let isNew = false;

  const eventStartTimes = {
    eu: 1637215200,
    com: 1637150400,
    ru: 1637128800,
    asia: 1637179200,
  };
  const ts = Math.round(+new Date() / 1000);
  const now = new Date(ts * 1000);

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
      } catch (e) {
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

<div class="font-sans w-full h-screen text-white bg-gray-900">
  <div
    class="relative w-full h-48 z-0"
    style="background: url(/header.jpg) no-repeat center center fixed; background-size: cover;"
  >
    <div
      class="absolute left-0 bottom-0 bg-gradient-to-b from-transparent to-gray-900 w-full h-full z-0"
    />
    <div class="flex justify-center p-4 pt-12 z-10">
      <div class="z-10 flex-shrink-0 hidden sm:block">
        <img
          alt="Logo made by AdonisWerther"
          class="h-24 w-auto"
          src="/img/christmas.png"
        />
      </div>
      <div class="z-10 pl-4">
        <h1 class="text-5xl text-white">Steelwhaling</h1>
        <div class="text-gray-200">
          Brought to you by Rukenshia on the EU server, the same idiot that
          built the other whaling websites and
          <a
            href="https://dashboard.twitch.tv/extensions/1n8nhpxd3p623wla18px8l8smy0ym7-2.2.1"
            style="color: inherit"
            class="font-medium underline">Shipvoting</a
          >
        </div>
      </div>
    </div>
  </div>

  <div class="w-3/4 2xl:w-1/2 mx-auto">
    <div class="mt-8 h-8 flex justify-between">
      <div class="flex items-center">
        <div class="text-gray-300 text-sm font-medium">
          {#if $loggedIn}
            {#if ts < eventStartTimes[$realm]}
              <a
                href="https://worldofwarships.ru/en/news/game-updates/update-0108-wows-anniversary/#wows-anniversary"
                class="hover:text-gray-100 flex items-center gap-2"
              >
                Event start: {formatRelative(
                  new Date(eventStartTimes[$realm] * 1000),
                  now
                )}
                <span class="text-gray-400 font-normal">(local time)</span>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                  />
                </svg>
              </a>
            {/if}
          {:else if ts < eventStartTimes['eu']}
            <a
              href="https://worldofwarships.ru/en/news/game-updates/update-0108-wows-anniversary/#wows-anniversary"
              class="hover:text-gray-100 flex items-center gap-2"
            >
              Event start: {formatRelative(
                new Date(eventStartTimes['eu'] * 1000),
                now
              )}
              <span class="text-gray-400 font-normal">(local time)</span>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                />
              </svg>
            </a>
          {/if}
        </div>
      </div>
      <div class="flex justify-end flex-wrap gap-2">
        <a
          href="https://www.buymeacoffee.com/rukenshia"
          target="_blank"
          class="px-4 font-xs font-medium border-none py-1 rounded
          bg-gray-50 hover:bg-gray-200 shadow-xl"
        >
          <div class="flex gap-2">
            <img
              src="/img/bmc-logo.svg"
              class="w-6 h-6"
              alt="buy me a coffee"
            />
            <span class="text-gray-900">Donate</span>
          </div>
        </a>
        <a
          on:click={contact}
          href="mailto:svc-frenchwhaling@ruken.pw"
          class="px-4 font-xs border-none py-1 rounded bg-gray-700
        hover:bg-gray-800"
        >
          Contact me
        </a>
        {#if $loggedIn}
          <button
            on:click={logout}
            class="px-4 font-xs border-none py-1 rounded bg-gray-700
          hover:bg-gray-800"
          >
            Logout
          </button>
        {/if}
      </div>
    </div>
  </div>
  <div class="bg-gray-900 sm:mt-0">
    <!--
    <div class="mt-12 w-full flex justify-around">
      <div class="flex flex-wrap mt-8 justify-around">
        <div class="w-full mx-auto">
          <div class="bg-gray-800 text-gray-400 rounded-sm text-md p-4">
            Sorry, this event is over.
          </div>
        </div>
      </div>
    </div>
    -->

    {#if $loggedIn}
      <div class="p-8 w-full flex justify-around">
        <div class="w-full xl:w-3/4">
          <div class="mt-4 mb-8 w-full justify-around flex">
            <div class="w-3/4 rounded p-2 bg-gray-800 text-white text-gray-300">
              If you enjoy using this website, please share the word and link
              your friends to
              <a
                href="https://whaling.in.fkn.space"
                style="color: inherit"
                class="font-bold"
              >
                https://whaling.in.fkn.space
              </a>
            </div>
          </div>

          <Progress {isNew} eventStarted={ts >= eventStartTimes[$realm]} />
        </div>
      </div>
    {:else}
      <div
        class="flex flex-wrap mt-10 gap-4 justify-center text-sm uppercase text-gray-400"
      >
        Global Progress
      </div>
      <div
        class="w-3/4 2xl:w-1/2 mx-auto grid grid-cols-3 gap-4 mt-2 justify-center"
      >
        {#each $statistics as res}
          <div
            class="bg-gray-800 text-gray-200 rounded-sm col-span-3 md:col-span-1"
          >
            <div class="p-4 flex gap-2 items-center">
              <div class="">
                <img
                  class="h-8 w-auto"
                  alt="resource"
                  src="/img/resources/{res.Type}.png"
                />
              </div>
              <div>
                {HRNumbers.toHumanString(res.Earned)}

                <span class="text-gray-400">
                  /
                  {HRNumbers.toHumanString(res.Amount)}
                </span>
              </div>
            </div>
            <div class="relative h-2 w-full z-0 bg-gray-700 rounded-b-sm">
              <div
                style="width: {(res.Earned / res.Amount) * 100}%"
                class="absolute bottom-0 h-2 bg-green-900 rounded-b-sm"
              />
            </div>
          </div>
        {/each}
      </div>
      <div class="flex flex-wrap mt-8 justify-around">
        {#if error}
          <div class="w-full flex justify-around mb-8">
            <div class="w-1/2 text-center bg-red-600 text-white rounded-sm p-4">
              There was an error logging you in. Feel free to contact me
              <a
                class="font-medium underline"
                href="mailto:svc-frenchwhaling@ruken.pw"
              >
                via Email
              </a>
              or Discord (Rukenshia#4396) if you can't get past this.
              <br />
              Error message:
              <span class="font-mono"> {reason} <span /> </span>
            </div>
          </div>
        {/if}

        <Login />
      </div>
      <div class="mb-16" />
    {/if}

    <div class="mt-8 mb-8 text-gray-400 font-medium text-sm text-center">
      <a href="#privacy" on:click={privacyPolicy}>Privacy Policy</a>
      &bullet;
      <a target="_blank" href="https://git.sr.ht/~rukenshia/frenchwhaling">
        Source code
      </a>
      &bullet; This website is not affiliated with Wargaming &bullet; Thanks to AdonisWerther
      for the logo ❤️
    </div>

    {#if toggle}
      <div class="flex justify-around">
        <div
          id="privacy"
          class="w-3/4 mb-8 pl-8 text-md text-gray-400 text-left"
        >
          <h3 class="text-3xl mb-4">Privacy Policy</h3>
          <p>
            To provide this service to you, the following data will be collected
            and stored:
          </p>

          <ul class="pl-8 mt-2">
            <li>Your Wargaming account id</li>
            <li>Statistics about ships you own in World of Warships</li>
            <li>
              Statistics about how you perform in battles (wins per game mode)
            </li>
          </ul>

          <p class="mt-2">
            The data will be stored on Amazon Web Services in the eu-central-1
            (Frankfurt) region. I intend to publish statistics about the event
            after it is over. Data will be anonymised and only general
            statistics will be provided to the public, meaning that published
            data will not be traceable to individuals.
          </p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  a:visited {
    @apply text-gray-400;
  }

  body {
    @apply text-white;
  }
</style>
