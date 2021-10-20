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

      sidebar.appendChild(document.createElement("hr"))
    } else {
      el.classList = "sidebar-item"
      collapsibleGroup.appendChild(el)
    }

    headingsLookup[heading.id] = currentHeadingGroup
  }

  let backgrounds = [document.getElementById("left-background"), document.getElementById("right-background")];
  for (let background of backgrounds) {
    for (let i = 0; i < 400; i++) {
      let dot = document.createElement("div")
      dot.classList = "dot"
      background.appendChild(dot)

      if (Math.random() < 0.05) {
        dot.classList.add("red")
      }
    }
  }

  function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min) + min); //The maximum is exclusive and the minimum is inclusive
  }

  setInterval(() => {
    let dots = document.querySelectorAll(".dot.red")
    for (let i = 0; i < dots.length/10; i++) {
      dots[getRandomInt(0, dots.length-1)].classList.toggle("red")
    }

    dots = document.querySelectorAll(".dot")
    for (let i = 0; i < dots.length/100; i++) {
      dots[getRandomInt(0, dots.length-1)].classList.toggle("red")
    }
  }, 2000)

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
        document.querySelectorAll("#sidebar .active").forEach(e => e.classList.remove("active"))
        document.getElementById(`${winner}-item`).classList.add("active")
        document.getElementById(`${headingsLookup[winner]}-item`).classList.add("active");

        ticking = false;
      });

      ticking = true;
    }
  })

  document.getElementById("disclosure").addEventListener("click", (e) => {
    document.getElementById("container").classList.toggle("reveal-sidebar")
  })

  document.getElementById("mask").addEventListener("click", (e) => {
    document.getElementById("container").classList.toggle("reveal-sidebar")
  })
})()
