/**
 * Home Assistant Blueprints - Interactive Features
 */

(function () {
  // Mobile menu toggle
  const mobileMenuBtn = document.querySelector('.mobile-menu-btn');
  const navLinks = document.querySelector('.nav-links');

  if (mobileMenuBtn && navLinks) {
    mobileMenuBtn.addEventListener('click', () => {
      navLinks.classList.toggle('active');
      mobileMenuBtn.classList.toggle('active');
    });

    // Close menu when clicking a link
    navLinks.querySelectorAll('a').forEach((link) => {
      link.addEventListener('click', () => {
        navLinks.classList.remove('active');
        mobileMenuBtn.classList.remove('active');
      });
    });
  }

  // Smooth scroll for anchor links
  document.querySelectorAll('a[href^="#"]').forEach((anchor) => {
    anchor.addEventListener('click', function (e) {
      e.preventDefault();
      const target = document.querySelector(this.getAttribute('href'));
      if (target) {
        const navHeight = document.querySelector('.navbar').offsetHeight;
        const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - navHeight - 20;

        window.scrollTo({
          top: targetPosition,
          behavior: 'smooth',
        });
      }
    });
  });

  // Reveal animations on scroll
  const revealElements = document.querySelectorAll('.feature-card, .blueprint-card, .step');

  const revealOnScroll = () => {
    const windowHeight = window.innerHeight;
    const revealPoint = 100;

    revealElements.forEach((element) => {
      const elementTop = element.getBoundingClientRect().top;

      if (elementTop < windowHeight - revealPoint) {
        element.classList.add('reveal', 'active');
      }
    });
  };

  // Debounce scroll handler
  let scrollTimeout;
  window.addEventListener('scroll', () => {
    if (scrollTimeout) {
      window.cancelAnimationFrame(scrollTimeout);
    }
    scrollTimeout = window.requestAnimationFrame(revealOnScroll);
  });

  // Initial check for elements already in view
  revealOnScroll();

  // Navbar background on scroll
  const navbar = document.querySelector('.navbar');
  const updateNavbar = () => {
    if (window.scrollY > 50) {
      navbar.style.background = 'rgba(3, 0, 20, 0.95)';
    } else {
      navbar.style.background = 'rgba(3, 0, 20, 0.8)';
    }
  };

  window.addEventListener('scroll', updateNavbar);
  updateNavbar();

  // Add copy-to-clipboard for import URLs (future enhancement)
  document.querySelectorAll('.btn-import').forEach((btn) => {
    btn.addEventListener('click', function (e) {
      // Track click for analytics if needed
      const blueprintName = this.closest('.blueprint-card').querySelector('h3').textContent;
      console.log(`Import clicked: ${blueprintName}`);
    });
  });

  // Typing effect for code preview (subtle enhancement)
  const codeContent = document.querySelector('.code-content code');
  if (codeContent) {
    codeContent.style.opacity = '0';
    setTimeout(() => {
      codeContent.style.transition = 'opacity 0.5s ease';
      codeContent.style.opacity = '1';
    }, 500);
  }

  // Parallax effect for hero section
  const heroVisual = document.querySelector('.hero-visual');
  if (heroVisual && window.innerWidth > 768) {
    window.addEventListener('scroll', () => {
      const scrolled = window.pageYOffset;
      const rate = scrolled * 0.3;
      heroVisual.style.transform = `translateY(${rate}px)`;
    });
  }

  // Add intersection observer for more performant animations
  if ('IntersectionObserver' in window) {
    const observerOptions = {
      root: null,
      rootMargin: '0px',
      threshold: 0.1,
    };

    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('reveal', 'active');
          observer.unobserve(entry.target);
        }
      });
    }, observerOptions);

    revealElements.forEach((el) => {
      el.classList.add('reveal');
      observer.observe(el);
    });
  }

  // Console easter egg
  console.log(
    '%c Home Assistant Blueprints ',
    'background: linear-gradient(135deg, #3b82f6, #8b5cf6); color: white; padding: 10px 20px; font-size: 16px; font-weight: bold; border-radius: 5px;'
  );
  console.log('Pro-grade automation templates for real homes');
  console.log('https://github.com/schoolboyqueue/home-assistant-blueprints');
})();
