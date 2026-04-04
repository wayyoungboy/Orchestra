// Orchestra Landing Page Scripts

document.addEventListener('DOMContentLoaded', () => {
    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Navbar background on scroll
    const navbar = document.querySelector('.navbar');
    window.addEventListener('scroll', () => {
        if (window.scrollY > 50) {
            navbar.style.background = 'rgba(15, 15, 20, 0.95)';
        } else {
            navbar.style.background = 'rgba(15, 15, 20, 0.8)';
        }
    });

    // Animate elements on scroll
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    document.querySelectorAll('.feature-card, .step').forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(20px)';
        el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(el);
    });

    // Terminal typing animation
    const typingElement = document.querySelector('.typing');
    if (typingElement) {
        const commands = ['claude --help', 'gemini chat', 'aider .', 'git status'];
        let commandIndex = 0;
        let charIndex = 0;
        let isDeleting = false;

        function typeCommand() {
            const currentCommand = commands[commandIndex];

            if (isDeleting) {
                typingElement.textContent = currentCommand.substring(0, charIndex - 1);
                charIndex--;
            } else {
                typingElement.textContent = currentCommand.substring(0, charIndex + 1);
                charIndex++;
            }

            let typeSpeed = isDeleting ? 50 : 100;

            if (!isDeleting && charIndex === currentCommand.length) {
                typeSpeed = 2000;
                isDeleting = true;
            } else if (isDeleting && charIndex === 0) {
                isDeleting = false;
                commandIndex = (commandIndex + 1) % commands.length;
                typeSpeed = 500;
            }

            setTimeout(typeCommand, typeSpeed);
        }

        setTimeout(typeCommand, 1000);
    }
});