/**
 * Extracts unique tags from button elements within the document,
 * sorts them alphabetically, and creates corresponding filter buttons.
 */
function initializeTagFilters() {
    // Create a Set to store unique tags.
    const tags = new Set();

    // Iterate through all button elements in the document.
    for (const tagElement of document.getElementsByTagName("button")) {
        // Add the trimmed inner HTML of each button to the tags Set.
        tagElement.innerHTML = tagElement.innerHTML.trim();
        tags.add(tagElement.innerHTML);
    }

    // Convert the Set of tags to an Array and sort it alphabetically.
    const sortedTags = Array.from(tags);
    sortedTags.sort(Intl.Collator().compare);

    // Get the container element where the tag buttons will be placed.
    const btnContainer = document.getElementById("buttons");

    // Create filter buttons for each tag.
    for (const tag of sortedTags) {
        const btn = document.createElement("button");
        btn.className = "on"; // Initially set the button to the 'on' state.
        btn.innerHTML = tag;
        btnContainer.appendChild(btn);
    }

    // Create a "Show All" button.
    const showAllBtn = document.createElement("button");
    showAllBtn.className = "on";
    showAllBtn.innerHTML = "⬤";
    showAllBtn.id = "show_all_btn";
    showAllBtn.title = "Select all tags";

    // Create a "Hide All" button.
    const hideAllBtn = document.createElement("button");
    hideAllBtn.className = "on";
    hideAllBtn.innerHTML = "⬤";
    hideAllBtn.id = "hide_all_btn";
    hideAllBtn.title = "De-select all tags";

    // Insert the "Show All" and "Hide All" buttons at the beginning of the container.
    btnContainer.insertBefore(hideAllBtn, btnContainer.firstChild);
    btnContainer.insertBefore(showAllBtn, btnContainer.firstChild);

    // Get all elements with the class 'detail' (the posts).
    const posts = document.getElementsByClassName('detail');

    // Create a Set to store the tag filter buttons (excluding "Show All" and "Hide All").
    const filterButtons = new Set();
    for (const btn of document.getElementsByTagName('button')) {
        if (btn.id !== "show_all_btn" && btn.id !== "hide_all_btn") {
            filterButtons.add(btn);
        }
    }

    /**
     * Refreshes the visibility of posts based on the currently active tag filters.
     */
    function refreshPosts() {
        for (const post of posts) {
            let isVisible = false;
            // Iterate through the tag buttons within each post.
            for (const btn of post.getElementsByTagName("button")) {
                // If any tag button within the post is in the 'on' state, show the post.
                if (btn.className === "on") {
                    post.style.display = "block";
                    break; // No need to check further tags for this post.
                }
                // If all tag buttons within the post are in the 'off' state, hide the post.
                post.style.display = "none";
            }
        }
    }

    // Add event listeners to the tag filter buttons.
    for (const btn of filterButtons) {
        btn.addEventListener("click", function (e) {
            // Check if all filter buttons were 'on' before the click.
            let allButtonsOn = true;
            for (const filterBtn of filterButtons) {
                if (filterBtn.className === "off") {
                    allButtonsOn = false;
                    break;
                }
            }

            const target = e.target;

            // If all buttons were on, turn all off except the clicked one.
            if (allButtonsOn) {
                for (const filterBtn of filterButtons) {
                    filterBtn.className = "off";
                }
                target.className = "on";
            } else {
                // Otherwise, toggle the state of the clicked button.
                target.className = target.className === "on" ? "off" : "on";
            }

            // Ensure consistency between the clicked button and other buttons with the same tag.
            const targetInner = target.innerHTML;
            for (const filterBtn of filterButtons) {
                if (targetInner === filterBtn.innerHTML) {
                    filterBtn.className = target.className;
                }
            }
            refreshPosts(); // Update the visibility of posts.
        }, false);
    }

    // Add event listener to the "Show All" button.
    showAllBtn.addEventListener("click", function (e) {
        for (const btn of filterButtons) {
            btn.className = "on"; // Set all filter buttons to 'on'.
        }
        refreshPosts(); // Update the visibility of posts.
    }, false);

    // Add event listener to the "Hide All" button.
    hideAllBtn.addEventListener("click", function (e) {
        for (const btn of filterButtons) {
            btn.className = "off"; // Set all filter buttons to 'off'.
        }
        refreshPosts(); // Update the visibility of posts.
    }, false);
}

// Call the function to initialize the tag filters when the script runs.
initializeTagFilters();