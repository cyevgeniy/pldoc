// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is responsible for "Quick jump" search
// box behaviour. It's inspired by the similar
// search box as in the Go language's package documentation
//
// The search box is shown by pressing the 'f' key.
// Arrow keys are used for moving between items in the list of
// available items on the page. Pressing the 'Enter' key
// performs jump to selected item. Pressing 'Escape' key
// closes the box, as well as clicking the 'Close' button.

// The root element of search box
let modal = document.getElementById('modal')

// Input field in the search box
let input = document.getElementById('searchBoxInput')

// Parent element for "jump to" items
let listWrapper = document.getElementById('list-wrap')

// List of the "jump to" items
var items = listWrapper.children

let activeItemIdx = -1

function isBoxVisible() {
    return !modal.classList.contains('invisible-item')
}

function showNode(node) {
    if (node) {
        node.classList.remove('invisible-item')
    }
}

function hideNode(node) {
    if (node) {
        node.classList.add('invisible-item')
    }
}

// Adoptation of the code from
// https://www.w3schools.com/howto/howto_js_filter_lists.asp
function filterItems(e) {

    if (e.key == 'ArrowDown' || e.key == 'ArrowUp') {
        return
    }

    // Declare variables
    var filter, txtValue;
    filter = input.value.toUpperCase();

    // Loop through all list items, and hide those who don't match the search query
    for (let item of items) {
        txtValue = item.textContent;

        if (txtValue.toUpperCase().indexOf(filter) > -1) {
            showNode(item)
        } else {
            hideNode(item)
        }
    }

    // On every new search, remove previous focused items
    activeItemIdx = -1;
    removeActiveItems();
    activateItem(activeItemIdx);
}

function makeAllItemsVisible() {
    if (!items) {
        return
    }

    for(let item of items) {
        showNode(item)
    }
}

function toggleDialog() {
    if (isBoxVisible()) {
        hideNode(modal)
    } else {
        showNode(modal)
        input.value = ""
        removeActiveItems()
        makeAllItemsVisible()
        input.focus()
    }
}

function keyDownHandler(e) {
    if ((e.key == 'Escape') && (isBoxVisible())) {
        hideNode(modal)
    }

    if (e.key == 'f' && !(e.ctrlKey || e.altKey || e.shiftKey || e.metaKey) && !isBoxVisible()) {
        toggleDialog()
        e.preventDefault()
    }

    if (!isBoxVisible()) {
        return
    }

    switch (e.key) {
    case 'ArrowDown':
        activateItem(activeItemIdx + 1)
        e.preventDefault()
        break
    case 'ArrowUp':
        activateItem(activeItemIdx - 1)
        e.preventDefault()
        break
    case 'Enter':
        let item = getActiveItem()
        if (item) {
            jumpToAnchor(item.getAttribute('href'))
        }
        toggleDialog()
        break
    }
}

function getActiveItem() {
    let items = getVisibleItems()

    if (!items) {
        return
    }

    if (activeItemIdx < items.length ) {
        return items[activeItemIdx]
    }
}

function jumpToAnchor(id) {
    location.hash = id
}

function getVisibleItems() {
    return listWrapper.querySelectorAll(':not(.invisible-item)')
}

function activateItem(num) {

    let list = getVisibleItems()

    if (!list) {
        return
    }

    // Remove 'active-item' class from the previous active item
    // if it differs from the activated item
    if (activeItemIdx != num && activeItemIdx >= 0) {
        list[activeItemIdx].classList.remove('active-item')
    }

    if (num >= 0 && num < list.length) {
        list[num].classList.add('active-item')

        // scroll if current active element is outside visible rectangle
        if (list[num].getBoundingClientRect().bottom > listWrapper.getBoundingClientRect().bottom) {
            list[num].scrollIntoView(false);
        }

        if (list[num].getBoundingClientRect().top < listWrapper.getBoundingClientRect().top) {
            list[num].scrollIntoView();
        }

        activeItemIdx = num
    }

    // Reset current item's index if activated index is
    // outside the list of visible items
    if (num > list.length - 1 || num < 0) {
        activeItemIdx = -1
    }
}

// Removes class 'active-item' from all items
// in the search box
function removeActiveItems() {
    if (!items) {
        return
    }

    for (let item of items) {
        item.classList.remove('active-item')
    }
}

function assignEventListeners() {
    document.addEventListener('keydown', keyDownHandler);

    if (input) {
        input.addEventListener('keyup', filterItems)
    }

    // Close search box when user clicks on an item
    for(let item of items){
        item.addEventListener("click", function(e) {
            toggleDialog();
        })
    }
}

assignEventListeners()
