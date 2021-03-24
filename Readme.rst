=========================================
Introduction
=========================================

**ParsinServer** is a Fingerprint-based Indoor Positioning System based on `Find <https://github.com/schollz/find>`_. In `Parsiot <https://parsiotco.ir>`_, we are working on Indoor Positioning Solutions and ParsinServer is one of the main products that we were working on. After developing the code for two years, in the end, we decided to make the code free and opensource. Our reasons are as follow:
    
* **Find is opensource** : Although we changed and refactor the code base, we used Find when we started. Therefore, we sense that as our duty to make the code opensource. 
* **Help other research** : As we spend a lot of our time to test and research on Fingerprint-based IPS solutions, We know how it can be useful or precious for any researcher to start their study and research on the pre-build code base.
* **Opportunity for new contributions** : Current IPS solutions that work with Wi-Fi and BLE are usually based on heuristic ideas. So we need others to contribute. 



.. attention::
    We don't provide documentation for ParsinServer and the code accessories(e.g. the Android and ios app). If you need more help email me (`Mohammad Hadi Azaddel, m.h.azaddel -- at -- gmail.com`).

.. _Quickstart Concepts:

Capabilities
===========================

**ParsinServer** contains complete capabilities as follow:



1. Main Dashboard
        
    .. image:: docs/images/d1.png 
        :alt: Dashboard
        
Dashboard consists of :
    * Algorithm configs 
    * Algorithm hyperparameter tuning
    * Display Algorithm details
    * Upload different map image
    * Total algorithm evaluation 
   
2. Collecting and Showing Fingerprints
           
    .. image:: docs/images/d2.png 
        :alt: Fingerprint Map
   


3. Live user positioning: 
    * User Management 
    * Collaboration with PDR solutions(on the android app)
    * Live UWB localization

    .. image:: docs/images/d3.png 
        :alt: Live map

4. Arbitrary fingerprint location:
           
    .. image:: docs/images/d4.png 
        :alt: Arbitrary Location

5. Access Points management:
           
    .. image:: docs/images/d5.png 
        :alt: access points locations

6. Fingerprint ambiguity: helps to find similar fingerprints
           
    .. image:: docs/images/d6.png 
        :alt: fingerprint ambiguity map   

7. Graph & Map: This tool provides:
    * Set map constraints and borders(It can be used in Particle Filter models)
    * Set valid connections between nodes
    
    .. image:: docs/images/d7.png 
        :alt: Map borders

    .. image:: docs/images/d8.png 
        :alt: Connection links

8. RSS heatmap

    .. image:: docs/images/d9.png 
        :alt: RSS heatmap

9. Error heatmap: 

    .. image:: docs/images/d10.png 
        :alt: Error heatmap

10. Offline tracking: Use to evaluate offline collected test data. E.g. in the below figure, you can see user track as a timeline.

    .. image:: docs/images/d11.png 
        :alt: Test track map

10. Error calculation: This is a complete panel to get different algorithm localization accuracy and localization details like the association between real fingerprint and the estimated positions.

    .. image:: docs/images/d12.png 
        :alt: Algorithm error calculation
    
11. Error comparison and CDF graph:

    .. image:: docs/images/d13.png 
        :alt: CDF1

    .. image:: docs/images/d14.png 
        :alt: CDF2

12. Particle Filter algorithm
13. UWB compatible data collection: You can collect fingerprints with the accurate localization systems like UWB. ## `Android and ios app <https://github.com/schollz/find>`_

Side Projects
===========================

14. `Tag tracking <https://github.com/ParsIOT/Parsin-rtls>`_ : You can use BLE and Wi-Fi tags and some signal receivers to track tags.
15. `Android app <https://github.com/ParsIOT/Find_BLE/>`_ and `react-native app (ios) <https://github.com/ParsIOT/RNative>`_
16. Proximity advertisement framework, `Parsin Admin Panel <https://github.com/ParsIOT/Parsin-Admin-Panel>`_. It works alongside Parsinserver. Besides we provided a proximity advertisement.
