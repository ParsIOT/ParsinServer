How To Write Documentation
==========================

Writing documentation
#####################
Using IDE
*********
Use a text editor and write your documentation in a ``.rst`` file.

you can install ``ReStructuredText Support`` plugin in JetBrains IDEs.

Using Microsoft Word
********************
You can write documentation in word and save them as ``html``, then use flowing command:

.. code-block:: bash

   pandoc -t rst -f WORDFILE.docx -o OUTPUTFILENAME.rst --reference-links

more information at `this <https://peintinger.com/?p=365>`_ tutorial

Syntax
######
.. tip::

    You can go to `ReadTheDocs <readthedocs.io>`_ and click on **Edit on GitHub** button in any documentation to see the source.

* `Official Documentation`_
* Here is a `CheatSheet`_ of ReStructured Text Syntax.
* Here is `Sphinx Documentation`_ of ReStructured Text Syntax.

.. Attention::

    You should to import your files in ``index.rst``

.. _Official Documentation: http://docutils.sourceforge.net/docs/ref/rst/restructuredtext.html
.. _CheatSheet: https://thomas-cokelaer.info/tutorials/sphinx/rest_syntax.html#internal-and-external-links
.. _Sphinx Documentation: http://www.sphinx-doc.org/en/stable/rest.html

Build The Documentation
#######################
every time you push your commits to gitlab, it rebuild the documentation, and it's accessible at `Project's Page`_

.. _Project's Page: http://parsiot.gitlab.io/ParsinServer

if you want to build documentation locally, you can run this commands in project root:

.. code-block:: bash

   cd Documentation
   make html

the open ``index.html`` in ``Documentation\_build\html\``.