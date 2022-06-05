    function myFunction(e) {

        if (e.key == 'ArrowDown' || e.key == 'ArrowUp') {
            return
        }
        
        // Declare variables
        var input, filter, ul, li, a, i, txtValue;
        input = document.getElementById('myInput');
        filter = input.value.toUpperCase();
        lw = document.getElementById('list-wrap');
        li = lw.getElementsByTagName('a');
        let match = false;

        // Loop through all list items, and hide those who don't match the search query
        for (i = 0; i < li.length; i++) {
            txtValue = li[i].textContent;

            if (txtValue.toUpperCase().indexOf(filter) > -1) {
                li[i].classList.remove('invisible-item')
            } else {
                li[i].classList.add('invisible-item')
            }
        }

        activeItemIdx = -1;
        removeActiveItems();
        activateItem(activeItemIdx);

    }

    function makeAllItemsVisible() {
        let itemList = document.getElementById('list-wrap')
        let list = itemList.children

        if (!list) {
            return
        }

        for(let i = 0; i < list.length; i++) {
            list[i].classList.remove('invisible-item')
        }
    }
        
    function toggleDialog() {
        modal = document.getElementById('modal')
        input = document.getElementById('myInput')

        if (!modal.classList.contains('invisible-item')) {
            modal.classList.add('invisible-item')
        } else {
            modal.classList.remove('invisible-item')
            input.value = ""
            removeActiveItems()
            makeAllItemsVisible()
            input.focus()
        }
    }

    function keyDownHandler(e) {
        modal = document.getElementById('modal')
        isVisible = !modal.classList.contains('invisible-item')
        if ((e.key == 'Escape') && (isVisible)) {
            modal.classList.add('invisible-item')
        }

        if (e.key == 'f' && !(e.ctrlKey || e.altKey || e.shiftKey || e.metaKey) && !isVisible) {
            toggleDialog()
            e.preventDefault()
        }

        if (e.key == 'ArrowDown' && isVisible) {
                                                
            activateItem(activeItemIdx + 1)
            e.preventDefault()
        }

        if (e.key == 'ArrowUp' && isVisible) {
            activateItem(activeItemIdx - 1)
            e.preventDefault()
        }

        if (e.key == 'Enter' && isVisible) {
            let item = getActiveItem()

            jumpToAnchor(item.getAttribute('href'))
            toggleDialog()
        }
    }

    let activeItemIdx = -1

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
        let itemList = document.getElementById('list-wrap')
        let list = itemList.querySelectorAll(':not(.invisible-item)')

        return list
    }
    
    function activateItem(num) {

        let list = getVisibleItems()

        if (!list) {
            return
        }
        
        if (activeItemIdx != num && activeItemIdx >= 0) {
            list[activeItemIdx].classList.remove('active-item')
        }
        
        if (num >= 0 && num <= list.length - 1) {
            list[num].classList.add('active-item')

            console.log(itemList.clientHeight)
            console.log(list[num].getBoundingClientRect().bottom)

            // scroll if current active element is outside visible rectangle
            if (list[num].getBoundingClientRect().bottom > itemList.getBoundingClientRect().bottom) {
                list[num].scrollIntoView(false);
            }

            if (list[num].getBoundingClientRect().top < itemList.getBoundingClientRect().top) {
                list[num].scrollIntoView();
            } 
            
            activeItemIdx = num
        }

        if (num > list.length - 1 || num < 0) {
            activeItemIdx = -1
        }
    }

    // Removes class 'active-item' from all items
    // in the search box
    function removeActiveItems() {
        let itemList = document.getElementById('list-wrap')
        let list = itemList.children;

        if (!list) {
            return
        }

        for (var i = 0; i < list.length; i++) {
            list[i].classList.remove('active-item')
        }
    }
    
    document.addEventListener('keydown', keyDownHandler);

    let input = document.getElementById('myInput')
    if (input) {
        input.addEventListener('keyup', myFunction)
    }

    let itemList = document.getElementById('list-wrap')
    let list = itemList.children

    // Close search box when user clicks on an item
    for(let i = 0; i < list.length; i++) {
        list[i].addEventListener("click", function(e) {
            toggleDialog();
        })
    }

