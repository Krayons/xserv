$(document).ready(function() {

    // Readable File sizes
    function humanFileSize(bytes, si) {
        var thresh = si ? 1000 : 1024;
        if(Math.abs(bytes) < thresh) {
            return bytes + ' B';
        }
            var units = si
                ? ['kB','MB','GB','TB','PB','EB','ZB','YB']
                : ['KiB','MiB','GiB','TiB','PiB','EiB','ZiB','YiB'];
            var u = -1;
            do {
                bytes /= thresh;
                ++u;
            } while(Math.abs(bytes) >= thresh && u < units.length - 1);
            return bytes.toFixed(1)+' '+units[u];
        }

    function epoch2date(epoch){
        var d = new Date(0); // The 0 there is the key, which sets the date to the epoch
        d.setUTCSeconds(epoch);
        return d
    }

    $("a[data-group='file']").each(function (){
        $(this).attr("href", "/downloads/" + location.pathname.substring(7) + $(this).text());
    });

    $("td[data-group='size']").each(function (){
        $(this).text(humanFileSize($(this).text(), true));
    });

    $("td[data-group='date']").each(function (){
        var nice_date = String(epoch2date($(this).text()));
        $(this).text(nice_date.substring(3, 15));
    });

    // Toggle Dark Theme
    $(".dark-toggle").click(function() {
        $("body").toggleClass("dark");
        if (localStorage.getItem("dark") == "on") {
            localStorage.setItem("dark", "off");
        }
        else {
            localStorage.setItem("dark", "on");
        }
    });

    if (typeof(Storage) !== "undefined") {
        if (localStorage.getItem("dark") === null) {
            localStorage.setItem("dark", "off");
        }
        if (localStorage.getItem("dark") === "on") {
            $(".dark-toggle").prop("checked", true);
            $("body").toggleClass("dark");
        }
    }
});
