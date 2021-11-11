(() => {
  let sidebar = document.getElementById("sidebar")
  let headings = document.querySelectorAll("#content h2,h3")
  let headingsLookup = {}
  let currentHeadingGroup = null
  let collapsibleGroup = null
  for (let heading of headings) {
    let el = document.createElement("a")
    el.href = `#${heading.id}`
    el.id = `${heading.id}-item`
    el.innerText = heading.innerText

    if (heading.tagName == "H2") {
      currentHeadingGroup = heading.id

      el.classList = "sidebar-heading"
      sidebar.appendChild(el)

      collapsibleGroup = document.createElement("div")
      collapsibleGroup.classList = "sidebar-section"
      sidebar.appendChild(collapsibleGroup)
    } else {
      el.classList = "sidebar-item"
      collapsibleGroup.appendChild(el)
    }

    headingsLookup[heading.id] = currentHeadingGroup
  }

  let ticking = false;
  document.getElementById("content").addEventListener("scroll", (e) => {
    if (!ticking) {
      window.requestAnimationFrame(function() {
        let viewportHeight = window.visualViewport.height;
        let winner = null;
        for (let heading of headings) {
          let rect = heading.getBoundingClientRect();
          if (rect.top > viewportHeight) { break }
          winner = heading.id
          if (rect.top > 0) { break }
        }
        if (winner != null) {
          document.querySelectorAll("#sidebar .active").forEach(e => e.classList.remove("active"))
          document.getElementById(`${winner}-item`).classList.add("active")
          document.getElementById(`${headingsLookup[winner]}-item`).classList.add("active");
        }
        ticking = false;
      });

      ticking = true;
    }
  })

  document.querySelector("img[alt=Ginkgo]").id = "top"

  document.querySelectorAll("div.highlight").forEach(el => {
    if (el.innerText.includes("/* === INVALID === */")) {
      el.classList.add("invalid")
    }
  })

  document.getElementById("disclosure").addEventListener("click", (e) => {
    document.getElementById("container").classList.toggle("reveal-sidebar")
  })

  document.getElementById("mask").addEventListener("click", (e) => {
    document.getElementById("container").classList.toggle("reveal-sidebar")
  })
})()
