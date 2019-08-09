
Algorithm structures:

Each algorithm must have two states:
    1.Learning:
        In this state, you must save learning data.
        Cache using :
         		gp := dbm.GM.GetGroup(groupName)
        		// Use gp to access learning data
        Algorithm learning data structure must be append in Group struct in groupcache.go
    2.Predicting
        Cache using :
            Two usage forms:
            1:(use it when many properties are needed and you want to set inner object properties line prop1.innerProp.aList[n])
             		gp := dbm.GM.GetGroup(groupName)
            2:
            		gp.Set_<property>(new value)
            		gp.Get_<property>()

finally add your algorithm to runner.go and crossValidation.go .
