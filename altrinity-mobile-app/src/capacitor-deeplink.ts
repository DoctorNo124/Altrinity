import { App, type URLOpenListenerEvent } from '@capacitor/app';

type UrlHandler = (url: string) => void;

export function initDeepLinkListener(onUrl: UrlHandler) {
  // 1️⃣  Listen for URLs while the app is already running
  App.addListener('appUrlOpen', (event: URLOpenListenerEvent) => {
    console.log('[DEEPLINK] live event →', event.url);
    onUrl(event.url);
  });

  // 2️⃣  Ask the native side what URL started the app
  App.getLaunchUrl().then(info => {
    if (info?.url) {
      console.log('[DEEPLINK] launch URL →', info.url);
      onUrl(info.url);
    } else {
      console.log('[DEEPLINK] no launch URL yet');
    }
  });
}
