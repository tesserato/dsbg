
var tags = new Set();
for (const tag_element of document.getElementsByTagName("button")) {
    tags.add(tag_element.innerHTML.trim());
}

const buttons = document.getElementsByTagName('button');

var btn_container = document.getElementById("buttons");

const show_all_btn = document.createElement("button");
show_all_btn.className = "on";
show_all_btn.innerHTML = "⬤";
show_all_btn.id = "show_all_btn";
show_all_btn.title = "Select all tags";

const hide_all_btn = document.createElement("button");
hide_all_btn.className = "on";
hide_all_btn.innerHTML = "⬤";
hide_all_btn.id = "hide_all_btn";
hide_all_btn.title = "De-select all tags";

for (const tag of tags) {
    var btn = document.createElement("button");
    btn.className = "on";
    btn.innerHTML = tag;
    btn_container.appendChild(btn);
}

btn_container.insertBefore(show_all_btn, btn_container.firstChild);
btn_container.insertBefore(hide_all_btn, btn_container.firstChild);
const posts = document.getElementsByClassName('detail');

function refresh_posts() {
    for (var post of posts) {
        for (var btn of post.getElementsByTagName("button")) {
            if (btn.className == "on") {
                post.style.display = "block";
                break;
            }
            post.style.display = "none";
        }
    }
}

for (const btn of buttons) {
    btn.addEventListener("click",
        function (e) {
            var target = e.target;
            target.className = target.className === "on" ? "off" : "on";
            for (var btn of buttons) {
                if (target.innerHTML.trim() == btn.innerHTML.trim()) {
                    btn.className = target.className;
                }
            }
            refresh_posts();
        }
        , false);
}

show_all_btn.addEventListener("click",
    function (e) {
        for (var btn of buttons) {
            btn.className = "on";
        }
        refresh_posts();
    }
    , false);

hide_all_btn.addEventListener("click",
    function (e) {
        for (var btn of buttons) {
            btn.className = "off";
        }
        refresh_posts();
    }
    , false);