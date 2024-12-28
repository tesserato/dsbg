var tags = new Set();
for (const tag_element of document.getElementsByTagName("button")) {
    tags.add(tag_element.innerHTML.trim());
}

var sortedTags = Array.from(tags);
sortedTags.sort(Intl.Collator().compare);

var btn_container = document.getElementById("buttons");


for (const tag of sortedTags) {
    console.log(tag);
    var btn = document.createElement("button");
    btn.className = "on";
    btn.innerHTML = tag;
    btn_container.appendChild(btn);
}

const posts = document.getElementsByClassName('detail');
const buttons = document.getElementsByTagName('button');

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

btn_container.insertBefore(hide_all_btn, btn_container.firstChild);
btn_container.insertBefore(show_all_btn, btn_container.firstChild);


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
            var all_buttons_on = true;
            for (var btn of buttons) {
                if (btn.className == "off") {
                    all_buttons_on = false;
                    break;
                }
            }
            var target = e.target;
            if (all_buttons_on) {
                for (var btn of buttons) {
                    btn.className = "off";
                }
                target.className = "on";
            } else {
                target.className = target.className === "on" ? "off" : "on";
            }

            const target_inner = target.innerHTML.trim();
            for (var btn of buttons) {
                if (target_inner == btn.innerHTML.trim()) {
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