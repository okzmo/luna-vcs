# LUNA

Luna is not a full blown VCS per say but a sugar coat on top of it you could say. It's a prototype I made to see if we could force a good workflow on people while still using git and honestly it works pretty well.

The idea is this:

You first init your repository, this will create a basic git repository, stage all your files and create an initial commit and also make a "luna" branch.
```bash
luna init
```


The luna branch is the long running branch of your project, similar to jujutsu you can think of the workflow as "branchless" even though it does use branches in the background. When you want to start working you must create a "workspace" otherwise you won't be able to work at all.

```bash
luna ws <name> <description>
```


This workspace is simply a branch behind the scene, and as you can see you give it a description, the description is here to explain what you're going to work on right now (e.g. "Adding X feature"), from there you can start working step by step.

When you want to add a step you do:
```bash
luna new <description>
```


This describe your next step which you're going to work on and everything you did before that will be commited for you with the description of your previous step, pretty cool right? Now when you're done with your work because you're an amazing SWE you simply do:
```bash
luna ws done
```

This take all your latest changes if any commit them and then squash everything and **rebase** that onto the **luna** branch with the description you've given at the beginning of it. Amazing no? A clean linear workflow. 

Of course it's pretty scarce in terms of features, there might be bugs but again it's a prototype to see if it was possible and clearly it is.

If you somehow want to sponsor this project so I can dive deeper on it or simply want to contribute hmu!
