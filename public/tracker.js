'use strict';

var queue = window.ps.q || [];
var trackerUrl = './tracker.png';
var website = '';
var commands = {
    "pageview": pageview,
    "create": create,
};

function stringifyObject(json) {
    var keys = Object.keys(json);

    keys = keys.filter(function(k) {
        return json[k].length > 0;
    });

    return '?' +
            keys.map(function(k) {
                    return encodeURIComponent(k) + '=' +
                            encodeURIComponent(json[k]);
            }).join('&');
}

function create(v) {
    website = v;
}

function pageview() {
    if( navigator.DonotTrack == 1 ) {
        return;
    }

    var path = location.pathname + location.search;
    var canonical = document.querySelector('link[rel="canonical"]');
    if(canonical && canonical.href) {
        path = canonical.href.substring(canonical.href.indexOf('/', 7)) || '/';
    }

    var d = {
        stamp: Date.now().toString(),
        w: website,
        h: location.hostname,
        t: document.title,
        l: navigator.language,
        p: path,
        s: screen.width + "x" + screen.height,
        r: document.referrer
    };

    var i = document.createElement('img');
    i.style.display = 'none';
    i.src = trackerUrl + stringifyObject(d);
    document.body.appendChild(i);
}

window.ps = function() {
    var args = [].slice.call(arguments);
    var c = args.shift();
    commands[c].apply(this, args);
};

queue.forEach(function(i) {
    ps.apply(this, i);
});