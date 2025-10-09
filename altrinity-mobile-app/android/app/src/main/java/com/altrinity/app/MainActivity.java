package com.altrinity.app;

import android.os.Bundle;
import android.webkit.WebView;

import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // ✅ Required to see the WebView in chrome://inspect
        WebView.setWebContentsDebuggingEnabled(true);
    }
}
